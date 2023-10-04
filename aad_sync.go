package itswizard_m_sync

import (
	"github.com/itslearninggermany/itswizard_m_msgraph"
	"github.com/jinzhu/gorm"
	"log"
)

type AadSyncCache struct {
	PersonToDelete          []DbPerson15                    `json:"person_to_delete"`
	PersonToDeleteExist     bool                            `json:"person_to_delete_exist"`
	PersonToImport          []DbPerson15                    `json:"person_to_import"`
	PersonToImportExist     bool                            `json:"person_to_import_exist"`
	PersonToUpdate          []PersonUpdate                  `json:"person_to_update"`
	PersonToUpdateExist     bool                            `json:"person_to_update_exist"`
	PersonsProblemsExist    bool                            `json:"persons_problems_exist"`
	PersonsProblems         []PersonProblem                 `json:"persons_problems"`
	MsrProblemExist         bool                            `json:"msr_problem_exist"`
	MsrProblem              []MsrProblem                    `json:"msr_problem"`
	MsrToDeleteExist        bool                            `json:"msr_to_delete_exist"`
	MsrToDelete             []DbMentorStudentRelationship15 `json:"msr_to_delete"`
	MsrToImportExist        bool                            `json:"msr_to_import_exist"`
	MsrToImport             []DbMentorStudentRelationship15 `json:"msr_to_import"`
	SprProblemsExist        bool                            `json:"spr_problems_exist"`
	SprProblem              []SprProblem                    `json:"spr_problem"`
	SprToDeleteExist        bool                            `json:"spr_to_delete_exist"`
	SprToDelete             []DbStudentParentRelationship15 `json:"spr_to_delete"`
	SprToImportExist        bool                            `json:"spr_to_import_exist"`
	SprToImport             []DbStudentParentRelationship15 `json:"spr_to_import"`
	MembershipProblemsExist bool                            `json:"membership_problems_exist"`
	MembershipProblems      []MembershipProblem             `json:"membership_problems"`
	MembershipToImportExist bool                            `json:"membership_to_import_exist"`
	MembershipToImport      []DbGroupMembership15           `json:"membership_to_import"`
	MembershipToDeleteExist bool                            `json:"membership_to_delete_exist"`
	MembershipToDelete      []DbGroupMembership15           `json:"membership_to_delete"`
	GroupsToImportExist     bool                            `json:"groups_to_import_exist"`
	GroupsToImport          []DbGroup15                     `json:"groups_to_import"`
	GroupsToDelete          []DbGroup15                     `json:"groups_to_delete"`
}

/*
Sync Method AAD
*/
func SyncMethodAAD(institution uint, organisation uint, personsFromImport []DbPerson15, msrFromImport []DbMentorStudentRelationship15, sprFromImport []DbStudentParentRelationship15, personsFromInstitutionDatabase []DbPerson15, personsFromOrganisationDatabase []DbPerson15, sprFromOrganisationDatabase []DbStudentParentRelationship15, msrFromOrganisationDatabase []DbMentorStudentRelationship15, membershipsFromDatabase []DbGroupMembership15, membershipFromImport []DbGroupMembership15, inputGroup []DbGroup15, groupsFromOrganisationDatabase []DbGroup15) (cache AadSyncCache) {
	log.Println("Check the Data if there are any contradictions.")
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
	log.Println("Finished check the Data if there are any contradictions.")

	log.Println("Start find out which persons are to delete")
	// 2. Find out, what is to delete.
	personsToDelete := getPersonToDelete(validPersons, personsFromOrganisationDatabase)

	for i := 0; i < len(personsToDelete); i++ {
		cache.PersonToDelete = append(cache.PersonToDelete, personsToDelete[i])
	}
	log.Println("Finished find out which persons are to delete")

	log.Println("Start finding out Memberships")
	membershipToImport, membershipToDelete, membershipProblems := syncMembership(membershipsFromDatabase, membershipFromImport, cache.PersonsProblems)

	for i := 0; i < len(membershipToImport); i++ {
		cache.MembershipToImport = append(cache.MembershipToImport, membershipToImport[i])
	}

	for i := 0; i < len(membershipProblems); i++ {
		cache.MembershipProblems = append(cache.MembershipProblems, membershipProblems[i])
	}

	for i := 0; i < len(membershipToDelete); i++ {
		cache.MembershipToDelete = append(cache.MembershipToDelete, membershipToDelete[i])
	}
	log.Println("Finish finding out Memberships")

	log.Println("Start finding out MSR")
	msrToImport, msrToDelete, msrProblem := syncMsr(msrFromOrganisationDatabase, validMsr, cache.PersonsProblems)
	for i := 0; i < len(msrToDelete); i++ {
		cache.MsrToDelete = append(cache.MsrToDelete, msrToDelete[i])
	}
	for i := 0; i < len(msrProblem); i++ {
		cache.MsrProblem = append(cache.MsrProblem, msrProblem[i])
	}
	for i := 0; i < len(msrToImport); i++ {
		cache.MsrToImport = append(cache.MsrToImport, msrToImport[i])
	}
	log.Println("Stop finding out MSR")

	log.Println("Start finding SPR")
	sprToImport, sprToDelete, sprProblem := syncSpr(sprFromOrganisationDatabase, validSpr, cache.PersonsProblems)
	for i := 0; i < len(sprToImport); i++ {
		cache.SprToImport = append(cache.SprToImport, sprToImport[i])
	}

	for i := 0; i < len(sprToDelete); i++ {
		cache.SprToDelete = append(cache.SprToDelete, sprToDelete[i])
	}
	for i := 0; i < len(sprProblem); i++ {
		cache.SprProblem = append(cache.SprProblem, sprProblem[i])
	}
	log.Println("Stop finding SPR")

	// 3. Import
	log.Println("Start finding Groups")
	groupsToDelete, groupsToImport := syncGroupAAD(institution, organisation, inputGroup, groupsFromOrganisationDatabase)
	for i := 0; i < len(groupsToImport); i++ {
		cache.GroupsToImport = append(cache.GroupsToImport, groupsToImport[i])
	}
	for i := 0; i < len(groupsToDelete); i++ {
		cache.GroupsToDelete = append(cache.GroupsToDelete, groupsToDelete[i])
	}
	log.Println("Stop finding Groups")

	// Persons SPR; MSR
	log.Println("Start to find Person to update or import")
	personsToImport, personsToUpdate, personsProblems := updateAndImportPersons(validPersons, personsFromInstitutionDatabase, personsFromOrganisationDatabase)
	for i := 0; i < len(personsProblems); i++ {
		cache.PersonsProblems = append(cache.PersonsProblems, personsProblems[i])
	}
	for i := 0; i < len(personsToImport); i++ {
		cache.PersonToImport = append(cache.PersonToImport, personsToImport[i])
	}
	for i := 0; i < len(personsToUpdate); i++ {
		cache.PersonToUpdate = append(cache.PersonToUpdate, personsToUpdate[i])
	}
	log.Println("Start to find Person to update or import")
	return
}

/*
Stores the Data from the Synccache to the Database, when there is an error no Rollback is done!!
*/
func (p *AadSyncCache) RunCache(aadCrawler, clientDb *gorm.DB, organisationID uint, institutionId uint) {
	log.Println("Start syncing with database")
	log.Println("Import Persons in Database")
	for i := 0; i < len(p.PersonToImport); i++ {
		err := clientDb.Save(&p.PersonToImport[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.PersonToImport[i].SyncPersonKey,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.PersonToImport[i].SyncPersonKey + " : " + p.PersonToImport[i].Username + " was added",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished importing Persons in Database")

	log.Println("Start updating Persons in Database")
	for i := 0; i < len(p.PersonToUpdate); i++ {
		err := clientDb.Save(&p.PersonToUpdate[i].Person).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.PersonToUpdate[i].Person.SyncPersonKey,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.PersonToUpdate[i].Person.SyncPersonKey + " : " + p.PersonToUpdate[i].Person.Username + " was updated",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished updating Persons in Database")

	log.Println("Start to delete Perons")
	for i := 0; i < len(p.PersonToDelete); i++ {
		err := clientDb.Delete(&p.PersonToDelete[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.PersonToDelete[i].SyncPersonKey,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.PersonToDelete[i].SyncPersonKey + " : " + p.PersonToDelete[i].Username + " was deleted",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished delete Persons.")

	log.Println("Start to import MSR")
	for i := 0; i < len(p.MsrToImport); i++ {
		err := clientDb.Save(&p.MsrToImport[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.MsrToImport[i].StudentSyncPersonKey + "::" + p.MsrToImport[i].MentorSyncPersonKey,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.MsrToImport[i].StudentSyncPersonKey + "::" + p.MsrToImport[i].MentorSyncPersonKey + " was added",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished to import MSR")

	log.Println("Start to delete MSR")
	for i := 0; i < len(p.MsrToDelete); i++ {
		err := clientDb.Delete(&p.MsrToDelete[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.MsrToDelete[i].StudentSyncPersonKey + "::" + p.MsrToDelete[i].MentorSyncPersonKey,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.MsrToDelete[i].StudentSyncPersonKey + "::" + p.MsrToDelete[i].MentorSyncPersonKey + " was deleted",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished to delete MSR")

	log.Println("Start to import SPR")
	for i := 0; i < len(p.SprToImport); i++ {
		err := clientDb.Save(&p.SprToImport[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.SprToImport[i].StudentSyncPersonKey + " ++ " + p.SprToImport[i].ParentSyncPersonKey,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.SprToImport[i].StudentSyncPersonKey + "::" + p.SprToImport[i].ParentSyncPersonKey + " was added",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished to import SPR")

	log.Println("Start to delete SPR")
	for i := 0; i < len(p.SprToDelete); i++ {
		err := clientDb.Delete(&p.SprToDelete[i]).Error
		if err != nil {
			log.Println("Error by Delete SPR, ", err.Error())
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.SprToDelete[i].StudentSyncPersonKey + " ++ " + p.SprToDelete[i].ParentSyncPersonKey,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving ErrorLog, ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.SprToDelete[i].StudentSyncPersonKey + "::" + p.SprToDelete[i].ParentSyncPersonKey + " was deleted",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished to delete SPR")

	log.Println("Start import Memberships")
	for i := 0; i < len(p.MembershipToImport); i++ {
		err := clientDb.Save(&p.MembershipToImport[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.MembershipToImport[i].PersonSyncKey + " ++ " + p.MembershipToImport[i].GroupName,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving ErrorLog, ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.MembershipToImport[i].PersonSyncKey + "::" + p.MembershipToImport[i].GroupName + " was added",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished import Memberships")

	log.Println("Start delete Memberships")
	for i := 0; i < len(p.MembershipToDelete); i++ {
		err := clientDb.Delete(&p.MembershipToDelete[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.MembershipToDelete[i].PersonSyncKey + " ++ " + p.MembershipToDelete[i].GroupName,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.MembershipToDelete[i].PersonSyncKey + "::" + p.MembershipToDelete[i].GroupName + " was deleted",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished delete Memberships")

	log.Println("Start import Groups")
	for i := 0; i < len(p.GroupsToImport); i++ {
		err := clientDb.Save(&p.GroupsToImport[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.GroupsToImport[i].Name,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.GroupsToImport[i].Name + " was imported",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished import Groups")

	log.Println("Start delete Groups")
	for i := 0; i < len(p.GroupsToDelete); i++ {
		err := clientDb.Delete(&p.GroupsToDelete[i]).Error
		if err != nil {
			err = aadCrawler.Save(&itswizard_m_msgraph.AadErrorLog{
				Type:           p.GroupsToDelete[i].Name,
				OrganisationID: organisationID,
				InstiutionID:   institutionId,
				Error:          err.Error(),
			}).Error
			if err != nil {
				log.Println("Error by saving AadErrorLog: ", err.Error())
			}
		}
		err = aadCrawler.Save(&itswizard_m_msgraph.AadLog{
			Type:           p.GroupsToDelete[i].Name + " was deleted",
			OrganisationID: organisationID,
			InstiutionID:   institutionId,
		}).Error
		if err != nil {
			log.Println("Error by saving AadLog: ", err.Error())
		}
	}
	log.Println("Finished delete Groups")
	log.Println("Finished syncing with database")
	return
}
