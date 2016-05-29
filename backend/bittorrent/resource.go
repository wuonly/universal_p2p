package bittorrent

import (
	"github.com/anacrolix/torrent"
	"time"
)

type Resource struct {
	lastAccess time.Time
	//for root
	Torrent      *torrent.Torrent
	SubFile      *torrent.File
	subResources []*Resource
}

func CreateFromTorrent(t *torrent.Torrent) *Resource {

	log.Debugf("create resource wrapper for torrent %s", t)

	root := &Resource{
		lastAccess: time.Now(),
		Torrent:    t,
	}

	root.subResources = make([]*Resource, len(t.Files()))

	for i, f := range t.Files() {

		log.Debugf("create sub resource for torrent %s,%s", t, f)
		root.subResources[i] = &Resource{
			//makes it old
			lastAccess: time.Unix(0, 0),
			SubFile:    &f,
		}
	}

	return root
}

func (this *Resource) Protocol() string {
	return PROTOCOL
}

func (this *Resource) Size() int64 {
	if this.IsRoot() {
		return this.Torrent.Length()
	}

	return this.SubFile.Length()
}

func (this *Resource) DiskUsage() int64 {
	//TODO
	return this.Size()
}

func (this *Resource) DownloadedSize() int64 {
	//TODO
	return this.Size()
}

func (this *Resource) LastAccess() time.Time {
	return this.lastAccess
}

func (this *Resource) IsRoot() bool {
	return this.Torrent != nil
}

func (this *Resource) RootURL() string {
	return this.Torrent.Info().Hash().HexString()
}

var emptyPathArray []string

//only sub resource have path
func (this *Resource) Path() []string {
	if this.IsRoot() {
		return emptyPathArray
	}

	return this.SubFile.FileInfo().Path
}
