package itswizard_m_sync

import "github.com/jinzhu/gorm"

type DbInstitution15 struct {
	gorm.Model
	Name         string
	CID          string `gorm:"unique"`
	ItslSiteName string `gorm:"unique"`
}

type DbOrganisation15 struct {
	gorm.Model
	Name          string
	InstitutionID uint `gorm:"TYPE:integer REFERENCES DbInstitution15"`
}

type DbMakeCourse struct {
	gorm.Model
	OrganisationID uint `gorm:"unique"`
	Course         bool
}

type DbSftpData15 struct {
	gorm.Model
	InstitutionId uint `gorm:"unique"`
	SftpUsername  string
	SftpPasswort  string
	SftpServer    string
}
