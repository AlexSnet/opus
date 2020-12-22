// Copyright © Go Opus Authors (see AUTHORS file)
//
// License for use of this code is detailed in the LICENSE file

package opus

/*
// Link opus using pkg-config.
#cgo pkg-config: opus
#include <opus.h>

#define MODE_SILK_ONLY          1000
#define MODE_HYBRID             1001
#define MODE_CELT_ONLY          1002

int bridge_decoder_get_packet_mode(const unsigned char *data) {
   int mode;
   if (data[0]&0x80) {
      mode = MODE_CELT_ONLY;
   } else if ((data[0]&0x60) == 0x60) {
      mode = MODE_HYBRID;
   } else {
      mode = MODE_SILK_ONLY;
   }
   return mode;
}

*/
import "C"

type Application int

const (
	// Optimize encoding for VoIP
	AppVoIP = Application(C.OPUS_APPLICATION_VOIP)
	// Optimize encoding for non-voice signals like music
	AppAudio = Application(C.OPUS_APPLICATION_AUDIO)
	// Optimize encoding for low latency applications
	AppRestrictedLowdelay = Application(C.OPUS_APPLICATION_RESTRICTED_LOWDELAY)
)

const (
	xMAX_BITRATE       = 48000
	xMAX_FRAME_SIZE_MS = 60
	xMAX_FRAME_SIZE    = xMAX_BITRATE * xMAX_FRAME_SIZE_MS / 1000
	// Maximum size of an encoded frame. I actually have no idea, but this
	// looks like it's big enough.
	maxEncodedFrameSize = 10000
)

func Version() string {
	return C.GoString(C.opus_get_version_string())
}

func GetFrameType(payload []byte) string {
	var frameType string
	switch payload[1]&0x3 {
	case 0:
		frameType = "One frame"
	case 1:
		frameType = "Two CBR frames"
	case 2:
		frameType = "Two VBR frames"
	case 3:
		frameType = "Multiple CBR/VBR frames (from 0 to 120 ms)"
	default:
	}

	return frameType
}

func GetPacketMode(payload []byte) int {
	return int(C.bridge_decoder_get_packet_mode((*C.uchar)(&payload[0])))
}