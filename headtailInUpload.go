package itswizard_m_sync

import "github.com/itslearninggermany/itswizard_m_userprovisioning"

func headTail(upload *upload) {
	upload.head = itswizard_m_userprovisioning.CreateXMLHead()
	upload.tail = itswizard_m_userprovisioning.CreateXMLTail()
}
