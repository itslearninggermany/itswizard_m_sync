package itswizard_m_sync

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/itslearninggermany/itswizard_m_objects"
)

// ///////////////////////////UPDATE/////////////////
func UpdateMethod(sessionUser itswizard_m_objects.SessionUser, personsFromOrganisationDatabase []DbPerson15, personFromInstitutionDatabase []DbPerson15, personsFromInput []DbPerson15, msrFromOrgnisationDatabase []DbMentorStudentRelationship15, membershipFromOrganisationDatabase []DbGroupMembership15, sprFromOrgnisationDatabase []DbStudentParentRelationship15, msrFromImport []DbMentorStudentRelationship15, sprFromImport []DbStudentParentRelationship15, membershipFromImport []DbGroupMembership15, groupsFromOrganisationDatabase []DbGroup15, groupsFromImport []DbGroup15, groupsFromInstitutionDatabase []DbGroup15) (cache SyncCache) {
	//Create Sets
	setPersonFromOrganisationDatabase := mapset.NewSet()
	setPersonFromInput := mapset.NewSet()
	setMsrFromOrgnisationDatabase := mapset.NewSet()
	setSprFromOrgnisationDatabase := mapset.NewSet()

	setMembershipFromOrganisationDatabase := mapset.NewSet()

	for i := 0; i < len(personsFromOrganisationDatabase); i++ {
		setPersonFromOrganisationDatabase.Add(personsFromOrganisationDatabase[i])
	}
	for i := 0; i < len(personsFromInput); i++ {
		setPersonFromInput.Add(personsFromInput[i])
	}
	for i := 0; i < len(msrFromOrgnisationDatabase); i++ {
		setMsrFromOrgnisationDatabase.Add(msrFromOrgnisationDatabase[i])
	}
	for i := 0; i < len(sprFromOrgnisationDatabase); i++ {
		setSprFromOrgnisationDatabase.Add(sprFromOrgnisationDatabase[i])
	}
	for i := 0; i < len(membershipFromOrganisationDatabase); i++ {
		setMembershipFromOrganisationDatabase.Add(membershipFromOrganisationDatabase[i])
	}

	personsToImport, personsToUpdate, personsProblems := updateAndImportPersons(personsFromInput, personFromInstitutionDatabase, personsFromOrganisationDatabase)
	for i := 0; i < len(personsProblems); i++ {
		cache.PersonsProblems = append(cache.PersonsProblems, personsProblems[i])
	}

	for i := 0; i < len(personsToImport); i++ {
		cache.PersonToImport = append(cache.PersonToImport, personsToImport[i])
	}

	for i := 0; i < len(personsToUpdate); i++ {
		cache.PersonToUpdate = append(cache.PersonToUpdate, personsToUpdate[i])
	}

	var syncIdsPersonProblems []string
	for i := 0; i < len(personsProblems); i++ {
		syncIdsPersonProblems = append(syncIdsPersonProblems, personsProblems[i].Person.SyncPersonKey)
	}

	for i := 0; i < len(msrFromImport); i++ {
		if existInList(msrFromImport[i].MentorSyncPersonKey, syncIdsPersonProblems) || existInList(msrFromImport[i].StudentSyncPersonKey, syncIdsPersonProblems) {
			cache.MsrProblem = append(cache.MsrProblem, MsrProblem{
				Problem: fmt.Sprint("Es gibt ein Poblem mit einer der Personen, siehe bitte Probleme bei Personen"),
				Msr:     msrFromImport[i],
			})
		} else {
			if setMsrFromOrgnisationDatabase.Contains(msrFromImport[i]) {
				cache.MsrProblem = append(cache.MsrProblem, MsrProblem{
					Problem: fmt.Sprint("Diese Verbindung besteht schon in der Datenbank"),
					Msr:     msrFromImport[i],
				})
			} else {
				cache.MsrToImport = append(cache.MsrToImport, msrFromImport[i])
			}
		}
	}

	for i := 0; i < len(sprFromImport); i++ {
		if existInList(sprFromImport[i].ParentSyncPersonKey, syncIdsPersonProblems) || existInList(sprFromImport[i].StudentSyncPersonKey, syncIdsPersonProblems) {
			cache.SprProblem = append(cache.SprProblem, SprProblem{
				Information: fmt.Sprint("Es gibt ein Poblem mit einer der Personen, siehe bitte Probleme bei Personen"),
				Spr:         sprFromImport[i],
			})
		} else {
			if setSprFromOrgnisationDatabase.Contains(sprFromImport[i]) {
				cache.SprProblem = append(cache.SprProblem, SprProblem{
					Information: fmt.Sprint("Diese Verbindung besteht schon in der Datenbank"),
					Spr:         sprFromImport[i],
				})
			} else {
				cache.SprToImport = append(cache.SprToImport, sprFromImport[i])
			}
		}
	}

	for i := 0; i < len(membershipFromImport); i++ {
		if existInList(membershipFromImport[i].PersonSyncKey, syncIdsPersonProblems) {
			cache.MembershipProblems = append(cache.MembershipProblems, MembershipProblem{
				Information: fmt.Sprint("Es gibt ein Poblem mit der Person: ", membershipFromImport[i].PersonSyncKey, "(siehe bitte Probleme bei Personen)"),
				Membership:  membershipFromImport[i],
			})
		} else {
			if setMembershipFromOrganisationDatabase.Contains(membershipFromImport[i]) {
				cache.MembershipProblems = append(cache.MembershipProblems, MembershipProblem{
					Information: fmt.Sprint("Diese Verbindung ist schon in der Datenbank"),
					Membership:  membershipFromImport[i],
				})
			} else {
				cache.MembershipToImport = append(cache.MembershipToImport, membershipFromImport[i])
			}
		}
	}

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
			cache.GroupsToImport = append(cache.GroupsToImport, DbGroup15{
				ID:                 fmt.Sprint(sessionUser.InstitutionID, "++", sessionUser.OrganisationID, "++", groupsFromImport[i].Name),
				SyncID:             groupsFromImport[i].Name,
				Name:               groupsFromImport[i].Name,
				ParentGroupID:      "rootPointer",
				Level:              1,
				DbInstitution15ID:  sessionUser.InstitutionID,
				DbOrganisation15ID: sessionUser.OrganisationID,
			})
		}
	}

	return
}

func existInList(name string, input []string) (exist bool) {
	for i := 0; i < len(input); i++ {
		if input[i] == name {
			return true
		}
	}
	return false
}
