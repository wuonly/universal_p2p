package bittorrent

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/juju/loggo"
	"github.com/zeronetscript/universal_p2p/backend"
	"github.com/zeronetscript/universal_p2p/backend/bittorrent"
	"net/http"
	"time"
)

var log = loggo.GetLogger("BittorrentFrontend")

type Frontend struct {
	backend *bittorrent.Backend
}

func Protocol(this *Frontend) string {
	return bittorrent.PROTOCOL
}

func SubVersion(this *Frontend) string {
	return "v0"
}

func getLargest(t *torrent.Torrent) *torrent.File {

	var target *torrent.File
	var maxSize int64
	for _, file := range t.Files() {
		if maxSize < file.Length() {
			maxSize = file.Length()
			target = &file
		}
	}

	return target
}

func pathEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (this *Frontend) Stream(w http.ResponseWriter,
	r *http.Request, access *backend.AccessRequest) {

	rootRes, err := this.backend.AddTorrentInfoHash(access.SubPath[0])

	if err != nil {
		log.Errorf(err.Error())
		http.Error(w, err.Error(), 404)
		return
	}

	var f *torrent.File

	if len(access.SubPath) == 1 {
		//ask for largest file in torrent
		f = getLargest(rootRes.Torrent)
	} else {
		have := false

		//TODO support archive unpack
		this.backend.IterateSubResources(rootRes, func(res backend.P2PResource) bool {
			cast := res.(*bittorrent.Resource)
			if pathEqual(access.SubPath, cast.SubFile.FileInfo().Path) {
				have = true
				f = cast.SubFile
				return true
			} else {
				return false
			}
		})
		if !have {
			errStr := fmt.Sprintf("no such file %s", access.SubPath)
			log.Errorf(errStr)
			http.Error(w, errStr, 404)
			return
		}
	}
	f.Download()

	reader, err := NewFileReader(f)

	defer func() {
		if err := reader.Close(); err != nil {
			log.Errorf("Error closing file reader: %s\n", err)
		}
	}()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+rootRes.Torrent.Info().Name+"\"")
	http.ServeContent(w, r, f.DisplayPath(), time.Now(), reader)

}

func HandleRequest(this *Frontend, w http.ResponseWriter, r *http.Request, request interface{}) {

	access := request.(*backend.AccessRequest)

	if len(access.SubPath) < 1 {
		log.Errorf("access url didn't have enough parameters")
		http.Error(w, "access url didin't have enough parameters", 404)
		return
	}

	if access.RootCommand == backend.STREAM {
		this.Stream(w, r, access)
		return
	} else {
		http.Error(w, "unsupport", http.StatusInternalServerError)
	}

}
