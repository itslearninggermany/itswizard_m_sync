package itswizard_m_sync

// ///////////////////////////DELETE/////////////////
func DeleteMethod(personsFromOrganisationDatabase []DbPerson15, personsFromInput []DbPerson15, msrFromDatabaseOrganisation []DbMentorStudentRelationship15, sprFromDatabaseOrganisation []DbStudentParentRelationship15, membershipsFromOrganisationDatabase []DbGroupMembership15) (cache SyncCache) {

	var personSyncKeysToDelete []string
	for i := 0; i < len(personsFromInput); i++ {
		personSyncKeysToDelete = append(personSyncKeysToDelete, personsFromInput[i].SyncPersonKey)
	}

	for i := 0; i < len(personSyncKeysToDelete); i++ {
		exist := false
		for x := 0; x < len(personsFromOrganisationDatabase); x++ {
			if personSyncKeysToDelete[i] == personsFromOrganisationDatabase[x].SyncPersonKey {
				cache.PersonToDelete = append(cache.PersonToDelete, personsFromOrganisationDatabase[x])
				exist = true
				break
			}
		}
		if !exist {
			for z := 0; z < len(personsFromInput); z++ {
				if personsFromInput[z].SyncPersonKey == personSyncKeysToDelete[i] {
					cache.PersonsProblems = append(cache.PersonsProblems, PersonProblem{
						Person:      personsFromInput[z],
						Information: "Die Person kann nicht gelöscht werden, da sie nicht in ihrer Organisation ist.",
					})
				}

			}
		}
	}

	/*

		//Create Sets
		setPersonFromOrganisationDatabase := mapset.NewSet()
		setPersonFromInput := mapset.NewSet()

		for i:= 0; i < len (personsFromOrganisationDatabase); i++ {
			setPersonFromOrganisationDatabase.Add(personsFromOrganisationDatabase[i])
		}
		for i:=0; i < len (personsFromInput);i ++ {
			setPersonFromInput.Add(personsFromInput[i])
		}

		//Get Problems
		intersection := setPersonFromInput.Intersect(setPersonFromOrganisationDatabase)
		setProblems := setPersonFromOrganisationDatabase.Difference(intersection)
		problems := setProblems.ToSlice()
		for i := 0; i < len (problems); i++ {
			personProblemData := problems[i].(DbPerson15)
			cache.PersonsProblems = append(cache.PersonsProblems, PersonProblem{
				Person:      personProblemData,
				Information: "Die Person kann nicht gelöscht werden, da sie sich nicht in der Organisation befindet.",
			})
		}


		// Get To Delete
		uncastPersonToDelete := intersection.ToSlice()
		for i:= 0; i < len (uncastPersonToDelete); i++ {
			cache.PersonToDelete = append(cache.PersonToDelete,uncastPersonToDelete[i].(DbPerson15))
		}



		var personSyncKeysToDelete []string
		for i := 0; i < len(cache.PersonToDelete); i++ {
			personSyncKeysToDelete = append(personSyncKeysToDelete, cache.PersonToDelete[i].SyncPersonKey)
		}
	*/
	//MSR
	for i := 0; i < len(msrFromDatabaseOrganisation); i++ {
		if existInList(msrFromDatabaseOrganisation[i].StudentSyncPersonKey, personSyncKeysToDelete) || existInList(msrFromDatabaseOrganisation[i].MentorSyncPersonKey, personSyncKeysToDelete) {
			cache.MsrToDelete = append(cache.MsrToDelete, msrFromDatabaseOrganisation[i])
		}
	}

	//SPR
	for i := 0; i < len(sprFromDatabaseOrganisation); i++ {
		if existInList(sprFromDatabaseOrganisation[i].StudentSyncPersonKey, personSyncKeysToDelete) || existInList(sprFromDatabaseOrganisation[i].ParentSyncPersonKey, personSyncKeysToDelete) {
			cache.SprToDelete = append(cache.SprToDelete, sprFromDatabaseOrganisation[i])
		}
	}

	//Memberships
	for i := 0; i < len(membershipsFromOrganisationDatabase); i++ {
		if existInList(membershipsFromOrganisationDatabase[i].PersonSyncKey, personSyncKeysToDelete) {
			cache.MembershipToDelete = append(cache.MembershipToDelete, membershipsFromOrganisationDatabase[i])
		}
	}

	return
}
