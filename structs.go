package itswizard_m_sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"strings"
)

// By inserting we have to proof if the Institution has already the username or the SyncPersonKey
type DbPerson15 struct {
	ID                 string `gorm:"not null; type:VARCHAR(500); primary key" json:"id"` // Example: DbInstitution15ID++DbOrganisation15ID++SyncPersonKey"+PersonIDNumber
	SyncPersonKey      string `gorm:"not null; type:VARCHAR(45);" json:"sync_person_key"`
	AzureUserID        string `gorm:"type:VARCHAR(150);" json:"azure_user_id"`
	FirstName          string `gorm:"not null; type:VARCHAR(45)" json:"first_name"`
	LastName           string `gorm:"not null; type:VARCHAR(45)" json:"last_name"`
	Username           string `gorm:"not null; type:VARCHAR(45)" json:"username"`
	Profile            string `gorm:"not null; type:VARCHAR(45)" json:"profile"`
	Password           string `gorm:"password; type:VARCHAR(45)" json:"password"`
	Email              string `gorm:"type:VARCHAR(45); primary key" json:"email"`
	Phone              string `gorm:"type:VARCHAR(45)" json:"phone"`
	Mobile             string `gorm:"type:VARCHAR(45)" json:"mobile"`
	Street1            string `gorm:"type:VARCHAR(45)" json:"street_1"`
	Street2            string `gorm:"type:VARCHAR(45)" json:"street_2"`
	Postcode           string `gorm:"type:VARCHAR(45)" json:"postcode"`
	City               string `gorm:"type:VARCHAR(45)" json:"city"`
	DbOrganisation15ID uint   `gorm:"TYPE:integer REFERENCES DbOrganisation15""`
	DbInstitution15ID  uint   `gorm:"TYPE:integer REFERENCES DbInstitution15" "`
}

type DbStudentParentRelationship15 struct {
	ID                   string `gorm:"primary key"` // Example: DbInstitution15ID++DbOrganisation15ID++StudentSyncPersonKey++ParentSyncPersonKey
	DbOrganisation15ID   uint   `gorm:"TYPE:integer REFERENCES DbOrganisation15"`
	DbInstitution15ID    uint   `gorm:"TYPE:integer REFERENCES DbInstitution15"`
	StudentSyncPersonKey string
	ParentSyncPersonKey  string
}

type DbMentorStudentRelationship15 struct {
	ID                   string `gorm:"primary key"` // Example: DbInstitution15ID++DbOrganisation15ID++MentorSyncPersonKey++StudentSyncPersonKey
	DbOrganisation15ID   uint   `gorm:"TYPE:integer REFERENCES DbOrganisation15"`
	DbInstitution15ID    uint   `gorm:"TYPE:integer REFERENCES DbInstitution15"`
	MentorSyncPersonKey  string
	StudentSyncPersonKey string
}

type DbGroup15 struct {
	ID                 string `gorm:"primary key"` //Example: DbInstitution15ID++DbOrganisation15ID++SyncID
	SyncID             string
	Name               string
	ParentGroupID      string // default: rootPointer
	Level              int    // default: 1
	IsCourse           bool
	DbInstitution15ID  uint `gorm:"TYPE:integer REFERENCES DbInstitution15"`
	DbOrganisation15ID uint `gorm:"TYPE:integer REFERENCES DbOrganisation15"`
}

type DbGroupMembership15 struct {
	ID                 string `gorm:"primary key"` //Example: DbInstitution15ID++DbOrganisation15ID++GroupName++PersonSyncKey
	PersonSyncKey      string
	GroupName          string
	DbInstitution15ID  uint `gorm:"TYPE:integer REFERENCES DbInstitution15"`
	DbOrganisation15ID uint `gorm:"TYPE:integer REFERENCES DbOrganisation15"`
}

type SyncCache struct {
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
}

type PersonProblem struct {
	Person      DbPerson15 `json:"person"`
	Information string     `json:"information"`
}
type PersonUpdate struct {
	Person      DbPerson15 `json:"person"`
	Information string     `json:"information"`
}
type MembershipProblem struct {
	Information string              `json:"information"`
	Membership  DbGroupMembership15 `json:"membership"`
}
type MsrProblem struct {
	Problem string                        `json:"problem"`
	Msr     DbMentorStudentRelationship15 `json:"msr"`
}
type SprProblem struct {
	Information string                        `json:"information"`
	Spr         DbStudentParentRelationship15 `json:"spr"`
}

/*
creates a JSON Object from the cache
*/
func (p *SyncCache) saveCacheInJson() (output string, outbyte []byte, err error) {
	outbyte, err = json.Marshal(&p)
	output = string(outbyte)
	return
}

/*
When there is a change it will return true
*/
func (p *SyncCache) Runable() bool {
	// persons
	if len(p.PersonToUpdate) > 0 {
		return true
	}
	if len(p.PersonToDelete) > 0 {
		return true
	}
	if len(p.PersonToImport) > 0 {
		return true
	}
	// groups
	if len(p.GroupsToImport) > 0 {
		return true
	}
	//Membership
	if len(p.MembershipToDelete) > 0 {
		return true
	}
	if len(p.MembershipToImport) > 0 {
		return true
	}
	// SPR
	if len(p.SprToDelete) > 0 {
		return true
	}
	if len(p.SprToImport) > 0 {
		return true
	}
	// MSR
	if len(p.MsrToDelete) > 0 {
		return true
	}
	if len(p.MsrToImport) > 0 {
		return true
	}
	return false
}

/*
Stores a cacheToTheDatabase
*/
func (p *SyncCache) SaveCacheToDatabase(insitutionID, organisationID uint, database *gorm.DB, service string) error {

	_, out, err := p.saveCacheInJson()
	if err != nil {
		return err
	}

	var errtmp error
	u1 := uuid.Must(uuid.NewV4(), errtmp)
	/*
		min := 0
		max := 1000000
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		newId := r.Intn(max-min) + min
	*/
	filename := fmt.Sprint(u1.String(), ".json")
	err = ioutil.WriteFile(filename, out, 666)
	if err != nil {
		return err
	}

	cont := DbSyncCache15{
		InstitutionID:  insitutionID,
		OrganisationID: organisationID,
		Content:        filename,
		Aad:            false,
		CSV:            false,
		Excel:          false,
		Rest:           false,
		Other:          false,
	}

	switch service {
	case "Aad":
		cont.Aad = true
	case "CSV":
		cont.CSV = true
	case "Excel":
		cont.Excel = true
	case "Rest":
		cont.Rest = true
	default:
		cont.Other = true
	}

	database.Save(&cont)
	return nil
}

/*
TODO: ....
*/
func (p *SyncCache) GetLast20CachesFromDatabase(insitutionID, organisationID uint) (out [20]SyncCache, err error) {
	return out, nil
}

/*
Gets a JSONObject and transfer it to a synccache
*/
func getCachefromJson(input string) (output SyncCache, err error) {
	err = json.Unmarshal([]byte(input), &output)
	return
}

/*
Stores the Data from the Synccache to the Database
*/
func (p *SyncCache) RunCache(database *gorm.DB) (importSuccessfull bool, log string, err error) {
	var errorColector []string
	tx := database.Begin()
	// All New Persons
	for i := 0; i < len(p.PersonToImport); i++ {
		err = tx.Save(&p.PersonToImport[i]).Error
		log = log + fmt.Sprintln(p.PersonToImport[i].SyncPersonKey, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.PersonToUpdate); i++ {
		err = tx.Save(&p.PersonToUpdate[i].Person).Error
		log = log + fmt.Sprintln(p.PersonToUpdate[i].Person.SyncPersonKey, " wurde upgedated")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.PersonToDelete); i++ {
		err = tx.Delete(&p.PersonToDelete[i]).Error
		log = log + fmt.Sprintln(p.PersonToDelete[i].SyncPersonKey, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MsrToImport); i++ {
		err = tx.Save(&p.MsrToImport[i]).Error
		log = log + fmt.Sprintln(p.MsrToImport[i].StudentSyncPersonKey, "::", p.MsrToImport[i].MentorSyncPersonKey, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MsrToDelete); i++ {
		err = tx.Delete(&p.MsrToDelete[i]).Error
		log = log + fmt.Sprintln(p.MsrToDelete[i].StudentSyncPersonKey, "::", p.MsrToDelete[i].MentorSyncPersonKey, " wurde gelöscht")

		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.SprToImport); i++ {
		err = tx.Save(&p.SprToImport[i]).Error
		log = log + fmt.Sprintln(p.SprToImport[i].StudentSyncPersonKey, "::", p.SprToImport[i].ParentSyncPersonKey, " wurde hinzugefügt")

		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.SprToDelete); i++ {
		err = tx.Delete(&p.SprToDelete[i]).Error
		log = log + fmt.Sprintln(p.SprToDelete[i].StudentSyncPersonKey, "::", p.SprToDelete[i].ParentSyncPersonKey, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.GroupsToImport); i++ {
		err = tx.Save(&p.GroupsToImport[i]).Error
		log = log + fmt.Sprintln(p.GroupsToImport[i].Name, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MembershipToImport); i++ {
		err = tx.Save(&p.MembershipToImport[i]).Error
		log = log + fmt.Sprintln(p.MembershipToImport[i].PersonSyncKey, "::", p.MembershipToImport[i].GroupName, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MembershipToDelete); i++ {
		err = tx.Delete(&p.MembershipToDelete[i]).Error
		log = log + fmt.Sprintln(p.MembershipToDelete[i].PersonSyncKey, "::", p.MembershipToDelete[i].GroupName, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	if len(errorColector) > 0 {
		importSuccessfull = false
		err = errors.New(strings.Join(errorColector, ""))
		// tx.Rollback()
	} else {
		importSuccessfull = true
		tx.Commit()
	}
	return
}

/*
Stores the Data from the Synccache to the Database, when there is an error no Rollback is done!!
*/
func (p *SyncCache) RunCacheWithoutRollback(database *gorm.DB) (log string, err error) {
	var errorColector []string
	// All New Persons
	for i := 0; i < len(p.PersonToImport); i++ {
		err = database.Save(&p.PersonToImport[i]).Error
		log = log + fmt.Sprintln(p.PersonToImport[i].SyncPersonKey, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.PersonToUpdate); i++ {
		err = database.Save(&p.PersonToUpdate[i].Person).Error
		log = log + fmt.Sprintln(p.PersonToUpdate[i].Person.SyncPersonKey, " wurde upgedated")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.PersonToDelete); i++ {
		err = database.Delete(&p.PersonToDelete[i]).Error
		log = log + fmt.Sprintln(p.PersonToDelete[i].SyncPersonKey, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MsrToImport); i++ {
		err = database.Save(&p.MsrToImport[i]).Error
		log = log + fmt.Sprintln(p.MsrToImport[i].StudentSyncPersonKey, "::", p.MsrToImport[i].MentorSyncPersonKey, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MsrToDelete); i++ {
		err = database.Delete(&p.MsrToDelete[i]).Error
		log = log + fmt.Sprintln(p.MsrToDelete[i].StudentSyncPersonKey, "::", p.MsrToDelete[i].MentorSyncPersonKey, " wurde gelöscht")

		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.SprToImport); i++ {
		err = database.Save(&p.SprToImport[i]).Error
		log = log + fmt.Sprintln(p.SprToImport[i].StudentSyncPersonKey, "::", p.SprToImport[i].ParentSyncPersonKey, " wurde hinzugefügt")

		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.SprToDelete); i++ {
		err = database.Delete(&p.SprToDelete[i]).Error
		log = log + fmt.Sprintln(p.SprToDelete[i].StudentSyncPersonKey, "::", p.SprToDelete[i].ParentSyncPersonKey, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.GroupsToImport); i++ {
		err = database.Save(&p.GroupsToImport[i]).Error
		log = log + fmt.Sprintln(p.GroupsToImport[i].Name, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MembershipToImport); i++ {
		err = database.Save(&p.MembershipToImport[i]).Error
		log = log + fmt.Sprintln(p.MembershipToImport[i].PersonSyncKey, "::", p.MembershipToImport[i].GroupName, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	for i := 0; i < len(p.MembershipToDelete); i++ {
		err = database.Delete(&p.MembershipToDelete[i]).Error
		log = log + fmt.Sprintln(p.MembershipToDelete[i].PersonSyncKey, "::", p.MembershipToDelete[i].GroupName, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint(err))
		}
	}

	if len(errorColector) > 0 {
		err = errors.New(strings.Join(errorColector, ""))
		// tx.Rollback()
	}

	return
}

type DbSyncCache15 struct {
	gorm.Model
	InstitutionID  uint
	OrganisationID uint
	Content        string `gorm:"type:MEDIUMTEXT"`
	Aad            bool
	CSV            bool
	Excel          bool
	Rest           bool
	Other          bool
}
