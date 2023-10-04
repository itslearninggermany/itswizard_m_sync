package itswizard_m_sync

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"log"
)

func SyncMethod(institution uint, organisation uint, personsFromImport []DbPerson15, msrFromImport []DbMentorStudentRelationship15, sprFromImport []DbStudentParentRelationship15, personsFromInstitutionDatabase []DbPerson15, personsFromOrganisationDatabase []DbPerson15, sprFromOrganisationDatabase []DbStudentParentRelationship15, msrFromOrganisationDatabase []DbMentorStudentRelationship15, membershipsFromDatabase []DbGroupMembership15, membershipFromImport []DbGroupMembership15, inputGroup []DbGroup15, groupsFromOrganisationDatabase []DbGroup15, groupsFromInstitutionDatabase []DbGroup15) (cache SyncCache) {
	// 1. Check the Data if there are any contradictions .
	validPersons, problems, validSpr, validMsr, problemSpr, problemMsr := makeFileValid(personsFromImport, sprFromImport, msrFromImport)

	for i := 0; i < len(problems); i++ {
		cache.PersonsProblems = append(cache.PersonsProblems, problems[i])
	}
	for i := 0; i < len(problemMsr); i++ {
		cache.MsrProblem = append(cache.MsrProblem, problemMsr[i])
	}

	for i := 0; i < len(problemSpr); i++ {
		cache.SprProblem = append(cache.SprProblem, problemSpr[i])
	}

	// 2. Find out, what is to delete.
	personsToDelete := getPersonToDelete(validPersons, personsFromOrganisationDatabase)

	for i := 0; i < len(personsToDelete); i++ {
		cache.PersonToDelete = append(cache.PersonToDelete, personsToDelete[i])
	}

	membershipToImport, membershipToDelete, membershipProblems := syncMembership(membershipsFromDatabase, membershipFromImport, cache.PersonsProblems)

	for i := 0; i < len(membershipProblems); i++ {
		cache.MembershipProblems = append(cache.MembershipProblems, membershipProblems[i])
	}

	for i := 0; i < len(membershipToDelete); i++ {
		cache.MembershipToDelete = append(cache.MembershipToDelete, membershipToDelete[i])
	}

	msrToImport, msrToDelete, msrProblem := syncMsr(msrFromOrganisationDatabase, validMsr, cache.PersonsProblems)
	for i := 0; i < len(msrToDelete); i++ {
		cache.MsrToDelete = append(cache.MsrToDelete, msrToDelete[i])
	}
	for i := 0; i < len(msrProblem); i++ {
		cache.MsrProblem = append(cache.MsrProblem, msrProblem[i])
	}

	sprToImport, sprToDelete, sprProblem := syncSpr(sprFromOrganisationDatabase, validSpr, cache.PersonsProblems)
	for i := 0; i < len(sprToDelete); i++ {
		cache.SprToDelete = append(cache.SprToDelete, sprToDelete[i])
	}
	for i := 0; i < len(sprProblem); i++ {
		cache.SprProblem = append(cache.SprProblem, sprProblem[i])
	}

	// 3. Import
	for i := 0; i < len(membershipToImport); i++ {
		cache.MembershipToImport = append(cache.MembershipToImport, membershipToImport[i])
	}

	groupsToImport := syncGroup(institution, organisation, inputGroup, groupsFromOrganisationDatabase, groupsFromInstitutionDatabase)
	for i := 0; i < len(groupsToImport); i++ {
		cache.GroupsToImport = append(cache.GroupsToImport, groupsToImport[i])
	}

	// Persons SPR; MSR
	personsToImport, personsToUpdate, personsProblems := updateAndImportPersons(validPersons, personsFromInstitutionDatabase, personsFromOrganisationDatabase)
	for i := 0; i < len(personsProblems); i++ {
		cache.PersonsProblems = append(cache.PersonsProblems, personsProblems[i])
	}

	for i := 0; i < len(personsToImport); i++ {
		cache.PersonToImport = append(cache.PersonToImport, personsToImport[i])
	}

	for i := 0; i < len(msrToImport); i++ {
		cache.MsrToImport = append(cache.MsrToImport, msrToImport[i])
	}

	for i := 0; i < len(sprToImport); i++ {
		cache.SprToImport = append(cache.SprToImport, sprToImport[i])
	}

	// 4. Update
	for i := 0; i < len(personsToUpdate); i++ {
		cache.PersonToUpdate = append(cache.PersonToUpdate, personsToUpdate[i])
	}

	return
}

/*
MEMBERSHIPS
*/
func syncMembership(membershipsFromOrganisationDatabase []DbGroupMembership15, membershipsFromImport []DbGroupMembership15, personProblemsFromFuncSyncPersons []PersonProblem) (membershipsToImport []DbGroupMembership15, membershipToDelete []DbGroupMembership15, membershipProblems []MembershipProblem) {
	// 0. Problemfälle auslesen und aus MembershipToImport löschen ==> MembershipProblems
	var personSyncIdsProblems []string
	for i := 0; i < len(personProblemsFromFuncSyncPersons); i++ {
		personSyncIdsProblems = append(personSyncIdsProblems, personProblemsFromFuncSyncPersons[i].Person.SyncPersonKey)
	}
	var membershipFromInputWithoutPersonProblems []DbGroupMembership15
	for i := 0; i < len(membershipsFromImport); i++ {
		if existInList(membershipsFromImport[i].PersonSyncKey, personSyncIdsProblems) {
			membershipProblems = append(membershipProblems, MembershipProblem{
				Information: fmt.Sprint("Die Zugehörigkeit für ", membershipsFromImport[i].PersonSyncKey, " zur Gruppe ", membershipsFromImport[i].GroupName, "kann nicht hinzugefügt werden, da ein Problem bei der Person vorliegt (siehe Probleme bei Personen)"),
				Membership:  membershipsFromImport[i],
			})
		} else {
			membershipFromInputWithoutPersonProblems = append(membershipFromInputWithoutPersonProblems, membershipsFromImport[i])
		}
	}
	// 1. Schnittmenge von MembershipToImport und MembershipInDatabse
	setMembershipsFromInputWithoutPersonProblems := mapset.NewSet()
	for i := 0; i < len(membershipFromInputWithoutPersonProblems); i++ {
		setMembershipsFromInputWithoutPersonProblems.Add(membershipFromInputWithoutPersonProblems[i])
	}

	setMembershipsFromOrganisationDatabase := mapset.NewSet()
	for i := 0; i < len(membershipsFromOrganisationDatabase); i++ {
		setMembershipsFromOrganisationDatabase.Add(membershipsFromOrganisationDatabase[i])
	}

	setIntersection := setMembershipsFromInputWithoutPersonProblems.Intersect(setMembershipsFromOrganisationDatabase)

	// 2. Schnitttmenge löschen aus MembershipFrom Import   ==> membershipsToImport
	tmp := setMembershipsFromOrganisationDatabase.Difference(setIntersection).ToSlice()

	for i := 0; i < len(tmp); i++ {
		tmp2, ok := tmp[i].(DbGroupMembership15)
		if ok {
			membershipToDelete = append(membershipToDelete, tmp2)
		} else {
			fmt.Println("Problem!!!")
		}
	}

	// 3. SChnittmenge löschen aus Membership in Database ==> MembershipToDelete
	tmp3 := setMembershipsFromInputWithoutPersonProblems.Difference(setIntersection).ToSlice()

	for i := 0; i < len(tmp3); i++ {
		tmp2, ok := tmp3[i].(DbGroupMembership15)
		if ok {
			membershipsToImport = append(membershipsToImport, tmp2)
		} else {
			fmt.Println("Problem!!!")
		}
	}
	return
}

/*
GROUPS
*/
func syncGroupAAD(institution uint, organisation uint, groupsFromImport []DbGroup15, groupsFromOrganisationDatabase []DbGroup15) (groupsToDelete, groupsToImport []DbGroup15) {
	igroup := mapset.NewSet()
	for i := 0; i < len(groupsFromImport); i++ {
		igroup.Add(groupsFromImport[i].SyncID)
	}
	dbgroup := mapset.NewSet()
	for i := 0; i < len(groupsFromOrganisationDatabase); i++ {
		dbgroup.Add(groupsFromOrganisationDatabase[i].SyncID)
	}
	groupsInBothSystems := dbgroup.Intersect(igroup)
	groupIDsForGroupsToDelete := dbgroup.Difference(groupsInBothSystems)
	groupIDsForGroupsToImport := igroup.Difference(groupsInBothSystems)
	//ToImport
	for i := 0; i < len(groupsFromImport); i++ {
		if groupIDsForGroupsToImport.Contains(groupsFromImport[i].SyncID) {
			groupsToImport = append(groupsToImport, groupsFromImport[i])
		}
	}
	//ToDelete
	for i := 0; i < len(groupsFromOrganisationDatabase); i++ {
		log.Println(groupsFromOrganisationDatabase[i].SyncID)
		if groupIDsForGroupsToDelete.Contains(groupsFromOrganisationDatabase[i].SyncID) {
			groupsToDelete = append(groupsToDelete, groupsFromOrganisationDatabase[i])
		}
	}
	return
}

/*
GROUPS
*/
func syncGroup(institution uint, organisation uint, groupsFromImport []DbGroup15, groupsFromOrganisationDatabase []DbGroup15, groupsFromInstitutionDatabase []DbGroup15) (groupsToImport []DbGroup15) {
	var groupnamesFromOrganisationDatabase []string
	var groupSyncKeysFromInstitutionDatabase []string

	for i := 0; i < len(groupsFromOrganisationDatabase); i++ {
		groupnamesFromOrganisationDatabase = append(groupnamesFromOrganisationDatabase, groupsFromOrganisationDatabase[i].Name)
	}
	for i := 0; i < len(groupsFromInstitutionDatabase); i++ {
		groupSyncKeysFromInstitutionDatabase = append(groupSyncKeysFromInstitutionDatabase, groupsFromInstitutionDatabase[i].SyncID)
	}

	for i := 0; i < len(groupsFromImport); i++ {
		if !existInList(groupsFromImport[i].Name, groupnamesFromOrganisationDatabase) {
			groupsToImport = append(groupsToImport, DbGroup15{
				ID:                 fmt.Sprint(institution, "++", organisation, "++", groupsFromImport[i].Name),
				SyncID:             groupsFromImport[i].Name,
				Name:               groupsFromImport[i].Name,
				ParentGroupID:      "rootPointer",
				Level:              1,
				DbInstitution15ID:  institution,
				DbOrganisation15ID: organisation,
			})
		}
	}

	return

}

/*
MSR
*/
func syncMsr(msrFromDatabaseOrganisation []DbMentorStudentRelationship15, msrFromImport []DbMentorStudentRelationship15, personProblems []PersonProblem) (msrToImport []DbMentorStudentRelationship15, msrToDelete []DbMentorStudentRelationship15, msrProblems []MsrProblem) {
	var personProblemSyncIds []string
	var msrFromImportWithoutProblems []DbMentorStudentRelationship15
	for i := 0; i < len(personProblems); i++ {
		personProblemSyncIds = append(personProblemSyncIds, personProblems[i].Person.SyncPersonKey)
	}
	for i := 0; i < len(msrFromImport); i++ {
		if existInList(msrFromImport[i].MentorSyncPersonKey, personProblemSyncIds) || existInList(msrFromImport[i].StudentSyncPersonKey, personProblemSyncIds) {
			msrProblems = append(msrProblems, MsrProblem{
				Problem: fmt.Sprint("Es gibt ein Problem mit einer der beiden Personen (siehe Probleme bei Personen)"),
				Msr:     msrFromImport[i],
			})
		} else {
			msrFromImportWithoutProblems = append(msrFromImportWithoutProblems, msrFromImport[i])
		}
	}

	setMsrFromImportWithoutProblems := mapset.NewSet()
	setMsrFromDatabaseOrganisation := mapset.NewSet()

	for i := 0; i < len(msrFromImportWithoutProblems); i++ {
		setMsrFromImportWithoutProblems.Add(msrFromImportWithoutProblems[i])
	}
	for i := 0; i < len(msrFromDatabaseOrganisation); i++ {
		setMsrFromDatabaseOrganisation.Add(msrFromDatabaseOrganisation[i])
	}

	setIntersection := setMsrFromDatabaseOrganisation.Intersect(setMsrFromImportWithoutProblems)
	tmp := setMsrFromDatabaseOrganisation.Difference(setIntersection).ToSlice()
	for i := 0; i < len(tmp); i++ {
		tmp2, ok := tmp[i].(DbMentorStudentRelationship15)
		if ok {
			msrToDelete = append(msrToDelete, tmp2)
		} else {
			fmt.Println("ERROR")
		}
	}

	tmp = setMsrFromImportWithoutProblems.Difference(setIntersection).ToSlice()
	for i := 0; i < len(tmp); i++ {
		tmp2, ok := tmp[i].(DbMentorStudentRelationship15)
		if ok {
			msrToImport = append(msrToImport, tmp2)
		} else {
			fmt.Println("ERROR")
		}
	}
	fmt.Println("Intersec")
	fmt.Println(setIntersection)

	return
}

/*
SPR
*/
func syncSpr(sprFromDatabaseOrganisation []DbStudentParentRelationship15, sprFromImport []DbStudentParentRelationship15, personProblems []PersonProblem) (sprToImport []DbStudentParentRelationship15, sprToDelete []DbStudentParentRelationship15, sprProblems []SprProblem) {
	var personProblemSyncIds []string
	var sprFromImportWithoutProblems []DbStudentParentRelationship15
	for i := 0; i < len(personProblems); i++ {
		personProblemSyncIds = append(personProblemSyncIds, personProblems[i].Person.SyncPersonKey)
	}
	for i := 0; i < len(sprFromImport); i++ {
		if existInList(sprFromImport[i].ParentSyncPersonKey, personProblemSyncIds) || existInList(sprFromImport[i].StudentSyncPersonKey, personProblemSyncIds) {
			sprProblems = append(sprProblems, SprProblem{
				Information: fmt.Sprint("Es gibt ein Problem mit einer der beiden Personen (siehe Probleme bei Personen)"),
				Spr:         sprFromImport[i],
			})
		} else {
			sprFromImportWithoutProblems = append(sprFromImportWithoutProblems, sprFromImport[i])
		}
	}

	setSprFromImportWithoutProblems := mapset.NewSet()
	setSprFromDatabaseOrganisation := mapset.NewSet()

	for i := 0; i < len(sprFromImportWithoutProblems); i++ {
		setSprFromImportWithoutProblems.Add(sprFromImportWithoutProblems[i])
	}
	for i := 0; i < len(sprFromDatabaseOrganisation); i++ {
		setSprFromDatabaseOrganisation.Add(sprFromDatabaseOrganisation[i])
	}

	setIntersection := setSprFromDatabaseOrganisation.Intersect(setSprFromImportWithoutProblems)
	tmp := setSprFromDatabaseOrganisation.Difference(setIntersection).ToSlice()
	for i := 0; i < len(tmp); i++ {
		tmp2, ok := tmp[i].(DbStudentParentRelationship15)
		if ok {
			sprToDelete = append(sprToDelete, tmp2)
		} else {
			fmt.Println("ERROR")
		}
	}
	tmp = setSprFromImportWithoutProblems.Difference(setIntersection).ToSlice()
	for i := 0; i < len(tmp); i++ {
		tmp2, ok := tmp[i].(DbStudentParentRelationship15)
		if ok {
			sprToImport = append(sprToImport, tmp2)
		} else {
			fmt.Println("ERROR")
		}
	}

	return
}

/*
PERSONS!!!
*/
func updateAndImportPersons(personsFromImport []DbPerson15, personsFromInstitutionDatabase []DbPerson15, personsFromOrganisationDatabase []DbPerson15) (personsToImport []DbPerson15, updatePersons []PersonUpdate, personProblems []PersonProblem) {
	var allPersonsFromOrganisationSyncIds []string
	var allPersonsFromOrganisationUsernames []string
	for i := 0; i < len(personsFromOrganisationDatabase); i++ {
		allPersonsFromOrganisationSyncIds = append(allPersonsFromOrganisationSyncIds, personsFromOrganisationDatabase[i].SyncPersonKey)
	}
	for i := 0; i < len(personsFromOrganisationDatabase); i++ {
		allPersonsFromOrganisationUsernames = append(allPersonsFromOrganisationUsernames, personsFromOrganisationDatabase[i].Username)
	}

	var allPersonsFromInstitutionSyncIds []string
	var allPersonsFromInstitutionUsernames []string
	for i := 0; i < len(personsFromInstitutionDatabase); i++ {
		allPersonsFromInstitutionSyncIds = append(allPersonsFromInstitutionSyncIds, personsFromInstitutionDatabase[i].SyncPersonKey)
		allPersonsFromInstitutionUsernames = append(allPersonsFromInstitutionUsernames, personsFromInstitutionDatabase[i].Username)
	}

	// Is the User part of the Organisation:
	for i := 0; i < len(personsFromImport); i++ {
		// Does the Person exist in the Organisation
		if existInList(personsFromImport[i].SyncPersonKey, allPersonsFromOrganisationSyncIds) {
			//Get Person from Organisation
			var personUpdate DbPerson15
			var information []string
			for a := 0; a < len(personsFromOrganisationDatabase); a++ {
				if personsFromOrganisationDatabase[a].SyncPersonKey == personsFromImport[i].SyncPersonKey {
					personUpdate = personsFromOrganisationDatabase[a]
					break
				}
			}

			if personUpdate.Username != personsFromImport[i].Username {
				information = append(information, " Der Nutzername wird geändert.")
				if existInList(personsFromImport[i].Username, allPersonsFromInstitutionUsernames) {
					personProblems = append(personProblems, PersonProblem{
						Person:      personsFromImport[i],
						Information: "Der Nutzername existiert schon, er kann nicht von " + personUpdate.Username + " zu " + personsFromImport[i].Username + " geändert werden",
					})
					continue
				}
			}

			if personUpdate.FirstName != personsFromImport[i].FirstName {
				information = append(information, " Der Vorname wird geändert.")
			}

			if personUpdate.LastName != personsFromImport[i].LastName {
				information = append(information, " Der Nachname wird geändert.")
			}

			if personUpdate.Email != personsFromImport[i].Email {
				information = append(information, " Die Email wird geändert.")
			}

			if personUpdate.Mobile != personsFromImport[i].Mobile {
				information = append(information, " Die Handynummer wird geändert.")
			}

			if personUpdate.Postcode != personsFromImport[i].Postcode {
				information = append(information, " Die Postleitzahl wird geändert.")
			}

			if personUpdate.City != personsFromImport[i].City {
				information = append(information, " Die Stadt wird geändert.")
			}

			if personUpdate.Phone != personsFromImport[i].Phone {
				information = append(information, " Die Telefonnummer wird geändert.")
			}

			if personUpdate.Profile != personsFromImport[i].Profile {
				information = append(information, " Das Profil wird geändert.")
			}

			if personUpdate.Street1 != personsFromImport[i].Street1 {
				information = append(information, " Die erste Adresszeile wird geändert.")
			}

			if personUpdate.Street2 != personsFromImport[i].Street2 {
				information = append(information, " Die zweite Adresszeile wird geändert.")
			}

			if personsFromImport[i].Password != "" {
				information = append(information, " Es wird ein neues Passwort vergeben.")
			}

			if len(information) > 0 {
				personsFromImport[i].ID = personUpdate.ID
				input := PersonUpdate{
					Person:      personsFromImport[i],
					Information: fmt.Sprint(information),
				}

				updatePersons = append(updatePersons, input)

			}
			continue
		}

		if existInList(personsFromImport[i].SyncPersonKey, allPersonsFromInstitutionSyncIds) {
			personProblems = append(personProblems, PersonProblem{
				Person:      personsFromImport[i],
				Information: "Die SyncID existiert schon in der Institution",
			})
			continue
		}
		if existInList(personsFromImport[i].Username, allPersonsFromInstitutionUsernames) {
			personProblems = append(personProblems, PersonProblem{
				Person:      personsFromImport[i],
				Information: "Der Username existiert schon in der Institution",
			})
			continue
		}

		// The Person is not in the Institution:
		personsToImport = append(personsToImport, personsFromImport[i])
	}
	return
}

func getPersonToDelete(persons []DbPerson15, fromDatabase []DbPerson15) (personsToDelete []DbPerson15) {
	for i := 0; i < len(fromDatabase); i++ {
		exist := false
		for in := 0; in < len(persons); in++ {
			if fromDatabase[i].SyncPersonKey == persons[in].SyncPersonKey || fromDatabase[i].Username == persons[in].Username {
				exist = true
				break
			}
		}
		if !exist {
			personsToDelete = append(personsToDelete, fromDatabase[i])
		}
	}
	return
}
