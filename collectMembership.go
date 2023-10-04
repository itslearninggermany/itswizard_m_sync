package itswizard_m_sync

import (
	"github.com/itslearninggermany/itswizard_m_objects"
	"github.com/itslearninggermany/itswizard_m_userprovisioning"
	"github.com/jinzhu/gorm"
)

func collectMemberships(dbClient *gorm.DB, user itswizard_m_objects.SessionUser, institutionName string, uploadData *upload) {

	var uploadcontent []byte

	// MEMBERSHIP:
	// Alle Gruppen der Institution
	var groups []DbGroup15
	err := dbClient.Where("db_institution15_id = ?", user.InstitutionID).Find(&groups).Error
	//Start Error-Handling
	//1. Log error
	if err != nil {
		uploadData.membershipError = err
		return
	}

	for i := 0; i < len(groups); i++ {
		// Hat die Gruppe Member?
		var groupMembership []DbGroupMembership15
		err = dbClient.Where("group_name = ?", groups[i].ID).Find(&groupMembership).Error
		//Start Error-Handling
		//1. Log error
		if err != nil {
			uploadData.membershipError = err
			return
		}

		if len(groupMembership) > 0 {
			tmp := itswizard_m_userprovisioning.Membership(institutionName, groups[i].ID)
			for g := 0; g < len(groupMembership); g++ {
				var person DbPerson15
				dbClient.Where("sync_person_key = ?", groupMembership[g].PersonSyncKey).Last(&person)
				tmp.AddMemeber(itswizard_m_userprovisioning.Member(institutionName, groupMembership[g].PersonSyncKey, person.Profile, false))
			}
			ergbyte, _, err := tmp.Parse2XML()
			//Start Error-Handling
			//1. Log error
			if err != nil {
				uploadData.membershipError = err
				return
			}

			for b := 0; b < len(ergbyte); b++ {
				uploadcontent = append(uploadcontent, ergbyte[b])
			}

		}

	}

	uploadData.memberships = uploadcontent
}
