package itswizard_m_sync

import (
	"fmt"
	"github.com/itslearninggermany/itswizard_m_objects"
	"github.com/itslearninggermany/itswizard_m_userprovisioning"
	"github.com/jinzhu/gorm"
	"strconv"
)

func collectGroups(dbClient *gorm.DB, user itswizard_m_objects.SessionUser, institutionName string, uploadData *upload) {

	var uploadcontent []byte
	// 2. Groups
	// Alle organisationen der instiutiotn bekommen

	var organisations []DbOrganisation15
	err := dbClient.Where("institution_id = ?", user.InstitutionID).Find(&organisations).Error
	//Start Error-Handling
	//1. Log error
	if err != nil {
		uploadData.groupError = err
		return
	}

	// Create the site
	group := itswizard_m_userprovisioning.Group(institutionName, "itslorganisation", "site", institutionName, "0", 0, "0", 0)
	ergbyte, _, err := group.Parse2XML()
	//Start Error-Handling
	//1. Log error
	if err != nil {
		uploadData.groupError = err
		return
	}

	for g := 0; g < len(ergbyte); g++ {
		uploadcontent = append(uploadcontent, ergbyte[g])
	}

	for i := 0; i < len(organisations); i++ {
		group := itswizard_m_userprovisioning.Group(institutionName, "itslorganisation", "school", organisations[i].Name, strconv.Itoa(int(organisations[i].ID)), 1, "0", 0)

		ergbyte, _, err := group.Parse2XML()
		//Start Error-Handling
		//1. Log error
		if err != nil {
			uploadData.groupError = err
			return
		}

		for g := 0; g < len(ergbyte); g++ {
			uploadcontent = append(uploadcontent, ergbyte[g])
		}

		//Get all groups
		var groups []DbGroup15
		err = dbClient.Where("db_institution15_id = ? AND db_organisation15_id = ?", user.InstitutionID, organisations[i].ID).Find(&groups).Error
		//Start Error-Handling
		//1. Log error
		if err != nil {
			uploadData.groupError = err
			return
		}

		for s := 0; s < len(groups); s++ {
			parentID := ""
			if groups[s].ParentGroupID == "rootPointer" {
				parentID = strconv.Itoa(int(organisations[i].ID))
			} else {
				parentID = groups[s].ParentGroupID
			}
			/*
				TODO: Abfrage ob mit Kurs oder ohne
			*/
			var makecourse DbMakeCourse
			err = dbClient.Where("organisation_id = ?", organisations[i].ID).First(&makecourse).Error
			if err != nil {
				uploadData.groupError = err
				return
			}

			fmt.Println(makecourse)
			if makecourse.Course {
				group = itswizard_m_userprovisioning.Group(institutionName, "", "COURSE", groups[s].Name, groups[s].ID, groups[s].Level, parentID, groups[s].Level-1)
			} else {
				if groups[s].IsCourse {
					group = itswizard_m_userprovisioning.Group(institutionName, "", "COURSE", groups[s].Name, groups[s].ID, groups[s].Level, parentID, groups[s].Level-1)
				} else {
					group = itswizard_m_userprovisioning.Group(institutionName, "", "", groups[s].Name, groups[s].ID, groups[s].Level, parentID, groups[s].Level-1)
				}

			}

			ergbyte, _, err := group.Parse2XML()
			//Start Error-Handling
			//1. Log error
			if err != nil {
				fmt.Println(err)
			}
			for g := 0; g < len(ergbyte); g++ {
				uploadcontent = append(uploadcontent, ergbyte[g])
			}

		}
	}
	uploadData.groups = uploadcontent
}
