package ws

import (
	"context"
	"testing"
	"time"

	"github.com/kiga-hub/arc/protocols"
	. "github.com/smartystreets/goconvey/convey"
)

func CaseService(t *testing.T) {
	// load config
	SetDefaultConfig()

	//初始化frame，组装DataGroup
	frame := createDataGroup()

	//Create and Start ws
	srv, err := New()
	if err != nil {
		t.Error("Fatal to Create: ", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go srv.Start(ctx)
	time.Sleep(time.Second * 3)

	Convey("CaseWrite: ", t, func() {
		if err := srv.Write(123, frame); err != nil {
			t.Error("Fatal to Write: ", err)
		}
	})
	cancel()
}

func TestService(t *testing.T) {
	t.Log("========TestService begin========")
	CaseService(t)
}

// 初始化frame，组装DataGroup
func createDataGroup() *protocols.Frame {
	frame := protocols.NewDefaultFrame()
	frame.Timestamp = time.Now().UnixNano() / 1e3
	frame.DataGroup = *protocols.NewDefaultDataGroup()
	frame.DataGroup.STypes = []byte{protocols.STypeArc}

	//STypeArc
	segmentAudio := protocols.NewDefaultSegmentArc()
	segmentAudio.Data = []byte{1, 2, 3}
	frame.DataGroup.Segments = append(frame.DataGroup.Segments, segmentAudio)

	return frame
}
