package itswizard_m_sync

import (
	"github.com/itslearninggermany/itswizard_m_objects"
	"github.com/jinzhu/gorm"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"sync"
)

type upload struct {
	persons         []byte
	personError     error
	groups          []byte
	groupError      error
	memberships     []byte
	membershipError error
	head            []byte
	tail            []byte
}

type FinalUploadStruct struct {
	UploadContent []byte
	UploadError   error
}

/*
With channels!! Make it faster!!
*/
func createTheUploadContent(dbClient *gorm.DB, user itswizard_m_objects.SessionUser, institutionName string) ([]byte, error) {
	var wg sync.WaitGroup
	var upload upload

	wg.Add(1)
	go collectPersons(dbClient, user, institutionName, &upload)
	wg.Add(1)
	go collectMemberships(dbClient, user, institutionName, &upload)
	wg.Add(1)
	go collectGroups(dbClient, user, institutionName, &upload)
	wg.Add(1)
	go headTail(&upload)
	wg.Wait()

	// check Errors
	if upload.personError != nil {
		return nil, upload.personError
	}
	if upload.groupError != nil {
		return nil, upload.groupError
	}
	if upload.membershipError != nil {
		return nil, upload.membershipError
	}

	//gwt all Datas
	output := upload.head
	for i := 0; i < len(upload.persons); i++ {
		output = append(output, upload.persons[i])
	}
	for i := 0; i < len(upload.groups); i++ {
		output = append(output, upload.groups[i])
	}
	for i := 0; i < len(upload.memberships); i++ {
		output = append(output, upload.memberships[i])
	}

	return output, nil

}

func SyncDataWithSFTP(user itswizard_m_objects.SessionUser, dbClient *gorm.DB, institutionName string, uploadStruct *FinalUploadStruct) {
	uploadcontent, err := createTheUploadContent(dbClient, user, institutionName)
	if err != nil {
		uploadStruct.UploadError = err
	}

	var sftpdata DbSftpData15
	err = dbClient.Where("institution_id = ?", user.InstitutionID).Last(&sftpdata).Error
	//Start Error-Handling
	//1. Log error
	if err != nil {
		uploadStruct.UploadError = err
	}

	config := &ssh.ClientConfig{
		User:            sftpdata.SftpUsername,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(sftpdata.SftpPasswort),
		},
	}

	config.SetDefaults()
	sshConn, err := ssh.Dial("tcp", sftpdata.SftpServer+":22", config) //sftpServer
	if err != nil {
		uploadStruct.UploadError = err
	}

	defer sshConn.Close()

	c, err := sftp.NewClient(sshConn)
	if err != nil {
		uploadStruct.UploadError = err
	}
	defer c.Close()

	// Uploading the file
	remoteFile, err := c.Create("upload.xml")
	if err != nil {
		uploadStruct.UploadError = err
	}

	_, err = remoteFile.Write(uploadcontent)
	if err != nil {
		uploadStruct.UploadError = err
	}

	uploadStruct.UploadContent = uploadcontent
}
