package itswizard_m_sync

/*
MAKE FILE VALID
*/
func makeFileValid(person []DbPerson15, spr []DbStudentParentRelationship15, msr []DbMentorStudentRelationship15) (personsWithoutProblems []DbPerson15, problems []PersonProblem, newSpr []DbStudentParentRelationship15, newMsr []DbMentorStudentRelationship15, sprProblems []SprProblem, msrProblems []MsrProblem) {
	problems, personsWithoutProblems = checkIfPersonsAreValid(person)
	newSpr, sprProblems = checkIfStudentParentRelationShipsAreValid(spr)
	newMsr, msrProblems = checkIfMentorStudentRelationShipsAreValid(msr)

	return
}

func checkIfPersonsAreValid(person []DbPerson15) (problems []PersonProblem, personsWithoutProblems []DbPerson15) {

	problem, person := makeFileValidPersonsProofEmptyUsernamesAndSyncKeys(person)
	problem1, person := makeFileValidPersonsProofDoubleSynckeys(person)
	problem2, personsWithoutProblems := makeFileValidPersonsProofDoubleUsernames(person)

	for i := 0; i < len(problem); i++ {
		problems = append(problems, problem[i])
	}

	for i := 0; i < len(problem1); i++ {
		problems = append(problems, problem1[i])
	}

	for i := 0; i < len(problem2); i++ {
		problems = append(problems, problem2[i])
	}

	return

}
func makeFileValidPersonsProofEmptyUsernamesAndSyncKeys(newPersons []DbPerson15) (problems []PersonProblem, personsWithoutProblems []DbPerson15) {
	// First: 	Check if there are Synckey double
	for i := 0; i < len(newPersons); i++ {
		if newPersons[i].Username == "" {
			problems = append(problems, PersonProblem{
				Person:      newPersons[i],
				Information: "Der Username ist leer",
			})
		}
		if newPersons[i].SyncPersonKey == "" {
			problems = append(problems, PersonProblem{
				Person:      newPersons[i],
				Information: "Der Synckey ist leer",
			})
		}
		if newPersons[i].SyncPersonKey != "" && newPersons[i].Username != "" {
			personsWithoutProblems = append(personsWithoutProblems, newPersons[i])
		}
	}
	return
}
func makeFileValidPersonsProofDoubleSynckeys(newPersons []DbPerson15) (problems []PersonProblem, personsWithoutProblems []DbPerson15) {
	// First: 	Check if there are Synckey double
	var allDoubleSyncIds []string
	for i := 0; i < len(newPersons); i++ {
		count := 0
		for in := 0; in < len(newPersons); in++ {
			if newPersons[i].SyncPersonKey == newPersons[in].SyncPersonKey {
				count++
			}
		}
		if count > 1 {
			allDoubleSyncIds = append(allDoubleSyncIds, newPersons[i].SyncPersonKey)
		}
	}

	// Take all the DoubleSyncIds Auccount out of newpersons
	for i := 0; i < len(newPersons); i++ {
		isDouble := false
		for in := 0; in < len(allDoubleSyncIds); in++ {
			if newPersons[i].SyncPersonKey == allDoubleSyncIds[in] {
				problems = append(problems, PersonProblem{
					Person:      newPersons[i],
					Information: "Die Sync Id ist im eingegebenen Datensatz mehrfach vergeben",
				})
				isDouble = true
				break
			}
		}
		if !isDouble {
			personsWithoutProblems = append(personsWithoutProblems, newPersons[i])
		}
	}

	return
}
func makeFileValidPersonsProofDoubleUsernames(newPersons []DbPerson15) (problems []PersonProblem, personsWithoutProblems []DbPerson15) {

	// First: 	Check if there are Synckey double
	var allDoubleUsernames []string
	for i := 0; i < len(newPersons); i++ {
		count := 0
		for in := 0; in < len(newPersons); in++ {
			if newPersons[i].Username == newPersons[in].Username {
				count++
			}
		}
		if count > 1 {
			allDoubleUsernames = append(allDoubleUsernames, newPersons[i].Username)
		}
	}

	// Take all the DoubleSyncIds Auccount out of newpersons
	for i := 0; i < len(newPersons); i++ {
		isDouble := false
		for in := 0; in < len(allDoubleUsernames); in++ {
			if newPersons[i].Username == allDoubleUsernames[in] {
				problems = append(problems, PersonProblem{
					Person:      newPersons[i],
					Information: "Der Username st im eingegebenen Datensatz mehrfach vergeben",
				})
				isDouble = true
				break
			}
		}
		if !isDouble {
			personsWithoutProblems = append(personsWithoutProblems, newPersons[i])
		}
	}

	return
}
func checkIfStudentParentRelationShipsAreValid(spr []DbStudentParentRelationship15) (newSpr []DbStudentParentRelationship15, sprProblems []SprProblem) {
	for i := 0; i < len(spr); i++ {
		if spr[i].StudentSyncPersonKey == spr[i].ParentSyncPersonKey {
			sprProblems = append(sprProblems, SprProblem{
				Information: "Datensatz wurde geändert, da die lernende Person mit dem SyncPersonKey " + spr[i].StudentSyncPersonKey + " nicht für sich selbst Erziehungsberechtigter sein kann.",
				Spr:         spr[i],
			})
		} else {
			newSpr = append(newSpr, spr[i])
		}
	}
	return
}
func checkIfMentorStudentRelationShipsAreValid(msr []DbMentorStudentRelationship15) (newMsr []DbMentorStudentRelationship15, msrProblem []MsrProblem) {
	for i := 0; i < len(msr); i++ {
		if msr[i].StudentSyncPersonKey == msr[i].MentorSyncPersonKey {
			msrProblem = append(msrProblem, MsrProblem{
				Problem: "Datensatz wurde geändert, da die Person mit dem SyncPersonKey " + msr[i].StudentSyncPersonKey + " nicht für sich selbst Mentor sein kann.",
				Msr:     msr[i],
			})
		} else {
			newMsr = append(newMsr, msr[i])
		}
	}
	return
}
