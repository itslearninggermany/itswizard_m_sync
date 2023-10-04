package itswizard_m_sync

import (
	"fmt"
	"github.com/itslearninggermany/itswizard_m_objects"
	"github.com/itslearninggermany/itswizard_m_userprovisioning"
	"github.com/jinzhu/gorm"
)

func collectPersons(dbClient *gorm.DB, user itswizard_m_objects.SessionUser, institutionName string, uploadData *upload) {
	// 1. Persons
	var persons []DbPerson15
	err := dbClient.Where("db_institution15_id = ?", user.InstitutionID).Find(&persons).Error
	//Start Error-Handling
	//1. Log error
	if err != nil {
		uploadData.personError = err
		return
	}

	for i := 0; i < len(persons); i++ {
		var studentparentrel []DbStudentParentRelationship15
		err = dbClient.Where("parent_sync_person_key = ?", persons[i].SyncPersonKey).Or("student_sync_person_key =?", persons[i].SyncPersonKey).Find(&studentparentrel).Error
		//Start Error-Handling
		//1. Log error
		if err != nil {
			uploadData.personError = err
			return
		}

		parent := false
		for g := 0; g < len(studentparentrel); g++ {
			if persons[i].SyncPersonKey == studentparentrel[g].ParentSyncPersonKey && persons[i].SyncPersonKey != studentparentrel[g].StudentSyncPersonKey {
				parent = true
				break
			}
		}
		var ids []string
		for s := 0; s < len(studentparentrel); s++ {
			if parent {
				ids = append(ids, studentparentrel[s].StudentSyncPersonKey)
			} else {
				ids = append(ids, studentparentrel[s].ParentSyncPersonKey)
			}

		}
		tel := itswizard_m_userprovisioning.MakeATelefonSlice([]int{1, 2}, []string{persons[i].Mobile, persons[i].Phone})
		p := itswizard_m_userprovisioning.Person(institutionName, persons[i].SyncPersonKey, persons[i].Username, persons[i].Password, persons[i].FirstName, persons[i].LastName, "", persons[i].Email, persons[i].Street1+persons[i].Street2, persons[i].City, persons[i].Postcode, persons[i].Profile, tel, ids, parent)
		ergbyte, ergstring, err := p.ParseToXML()
		//Start Error-Handling
		//1. Log error
		if err != nil {
			uploadData.personError = err
			return
		}

		var uploadcontent []byte
		for i := 0; i < len(ergbyte); i++ {
			uploadcontent = append(uploadcontent, ergbyte[i])
		}
		uploadData.persons = uploadcontent
		fmt.Println(ergstring)
	}
}
