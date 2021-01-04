// Copyright © Go Opus Authors (see AUTHORS file)
//
// License for use of this code is detailed in the LICENSE file

package opus

import (
	"fmt"
	"unsafe"
)

/*
#cgo pkg-config: opus
#include <opus.h>

int bridge_encoder_set_dtx(OpusEncoder *st, int use_dtx) {
	return opus_encoder_ctl(st, OPUS_SET_DTX(use_dtx));
}

int bridge_encoder_get_dtx(OpusEncoder *st, int *dtx) {
	return opus_encoder_ctl(st, OPUS_GET_DTX(dtx));
}

int bridge_encoder_set_vbr(OpusEncoder *st, int enable_vbr) {
	return opus_encoder_ctl(st, OPUS_SET_VBR(enable_vbr));
}

int bridge_encoder_get_vbr(OpusEncoder *st, int *vbr) {
	return opus_encoder_ctl(st, OPUS_GET_VBR(vbr));
}

int bridge_encoder_set_vbr_constraint(OpusEncoder *st, int vbr_constraint) {
	return opus_encoder_ctl(st, OPUS_SET_VBR_CONSTRAINT(vbr_constraint));
}

int bridge_encoder_get_vbr_constraint(OpusEncoder *st, int *vbr_constraint) {
	return opus_encoder_ctl(st, OPUS_GET_VBR_CONSTRAINT(vbr_constraint));
}

int bridge_encoder_set_signal(OpusEncoder *st, int signal_type) {
	return opus_encoder_ctl(st, OPUS_SET_SIGNAL(signal_type));
}

int bridge_encoder_get_signal(OpusEncoder *st, int *signal_type) {
	return opus_encoder_ctl(st, OPUS_GET_SIGNAL(signal_type));
}

int bridge_encoder_get_sample_rate(OpusEncoder *st, opus_int32 *sample_rate) {
	return opus_encoder_ctl(st, OPUS_GET_SAMPLE_RATE(sample_rate));
}

int bridge_encoder_set_bitrate(OpusEncoder *st, opus_int32 bitrate) {
	return opus_encoder_ctl(st, OPUS_SET_BITRATE(bitrate));
}

int bridge_encoder_get_bitrate(OpusEncoder *st, opus_int32 *bitrate) {
	return opus_encoder_ctl(st, OPUS_GET_BITRATE(bitrate));
}

int bridge_encoder_set_complexity(OpusEncoder *st, int complexity) {
	return opus_encoder_ctl(st, OPUS_SET_COMPLEXITY(complexity));
}

int bridge_encoder_get_complexity(OpusEncoder *st, int *complexity) {
	return opus_encoder_ctl(st, OPUS_GET_COMPLEXITY(complexity));
}

int bridge_encoder_set_bandwidth(OpusEncoder *st, int bandwidth) {
	return opus_encoder_ctl(st, OPUS_SET_BANDWIDTH(bandwidth));
}

int bridge_encoder_get_bandwidth(OpusEncoder *st, int *bandwidth) {
	return opus_encoder_ctl(st, OPUS_GET_BANDWIDTH(bandwidth));
}

int bridge_encoder_set_max_bandwidth(OpusEncoder *st, int max_bw) {
	return opus_encoder_ctl(st, OPUS_SET_MAX_BANDWIDTH(max_bw));
}

int bridge_encoder_get_max_bandwidth(OpusEncoder *st, int *max_bw) {
	return opus_encoder_ctl(st, OPUS_GET_MAX_BANDWIDTH(max_bw));
}

int bridge_encoder_set_inband_fec(OpusEncoder *st, int fec) {
	return opus_encoder_ctl(st, OPUS_SET_INBAND_FEC(fec));
}

int bridge_encoder_get_inband_fec(OpusEncoder *st, int *fec) {
	return opus_encoder_ctl(st, OPUS_GET_INBAND_FEC(fec));
}

int bridge_encoder_set_packet_loss_perc(OpusEncoder *st, int loss_perc) {
	return opus_encoder_ctl(st, OPUS_SET_PACKET_LOSS_PERC(loss_perc));
}

int bridge_encoder_get_packet_loss_perc(OpusEncoder *st, int *loss_perc) {
	return opus_encoder_ctl(st, OPUS_GET_PACKET_LOSS_PERC(loss_perc));
}

int bridge_encoder_set_force_channel(OpusEncoder *st, size_t num_channels) {
  if (num_channels == 0) {
    return opus_encoder_ctl(st, OPUS_SET_FORCE_CHANNELS(OPUS_AUTO));
  } else if (num_channels == 1 || num_channels == 2) {
    return opus_encoder_ctl(st, OPUS_SET_FORCE_CHANNELS(num_channels));
  } else {
    return -1;
  }
}

*/
import "C"

type Bandwidth int

const (
	// 4 kHz passband
	Narrowband = Bandwidth(C.OPUS_BANDWIDTH_NARROWBAND)
	// 6 kHz passband
	Mediumband = Bandwidth(C.OPUS_BANDWIDTH_MEDIUMBAND)
	// 8 kHz passband
	Wideband = Bandwidth(C.OPUS_BANDWIDTH_WIDEBAND)
	// 12 kHz passband
	SuperWideband = Bandwidth(C.OPUS_BANDWIDTH_SUPERWIDEBAND)
	// 20 kHz passband
	Fullband = Bandwidth(C.OPUS_BANDWIDTH_FULLBAND)
)

var errEncUninitialized = fmt.Errorf("opus encoder uninitialized")

// Encoder contains the state of an Opus encoder for libopus.
type Encoder struct {
	p        *C.struct_OpusEncoder
	channels int
	// Memory for the encoder struct allocated on the Go heap to allow Go GC to
	// manage it (and obviate need to free())
	mem []byte
}

// NewEncoder allocates a new Opus encoder and initializes it with the
// appropriate parameters. All related memory is managed by the Go GC.
func NewEncoder(sample_rate int, channels int, application Application) (*Encoder, error) {
	var enc Encoder
	err := enc.Init(sample_rate, channels, application)
	if err != nil {
		return nil, err
	}
	return &enc, nil
}

// Init initializes a pre-allocated opus encoder. Unless the encoder has been
// created using NewEncoder, this method must be called exactly once in the
// life-time of this object, before calling any other methods.
func (enc *Encoder) Init(sample_rate int, channels int, application Application) error {
	if enc.p != nil {
		return fmt.Errorf("opus encoder already initialized")
	}
	if channels != 1 && channels != 2 {
		return fmt.Errorf("Number of channels must be 1 or 2: %d", channels)
	}
	size := C.opus_encoder_get_size(C.int(channels))
	enc.channels = channels
	enc.mem = make([]byte, size)
	enc.p = (*C.OpusEncoder)(unsafe.Pointer(&enc.mem[0]))
	errno := int(C.opus_encoder_init(
		enc.p,
		C.opus_int32(sample_rate),
		C.int(channels),
		C.int(application)))
	if errno != 0 {
		return Error(int(errno))
	}
	return nil
}

// Encode raw PCM data and store the result in the supplied buffer. On success,
// returns the number of bytes used up by the encoded data.
func (enc *Encoder) Encode(pcm []int16, data []byte) (int, error) {
	if enc.p == nil {
		return 0, errEncUninitialized
	}
	if len(pcm) == 0 {
		return 0, fmt.Errorf("opus: no data supplied")
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("opus: no target buffer")
	}
	// libopus talks about samples as 1 sample containing multiple channels. So
	// e.g. 20 samples of 2-channel data is actually 40 raw data points.
	if len(pcm)%enc.channels != 0 {
		return 0, fmt.Errorf("opus: input buffer length must be multiple of channels")
	}
	samples := len(pcm) / enc.channels
	n := int(C.opus_encode(
		enc.p,
		(*C.opus_int16)(&pcm[0]),
		C.int(samples),
		(*C.uchar)(&data[0]),
		C.opus_int32(cap(data))))
	if n < 0 {
		return 0, Error(n)
	}
	return n, nil
}

// Encode raw PCM data and store the result in the supplied buffer. On success,
// returns the number of bytes used up by the encoded data.
func (enc *Encoder) EncodeFloat32(pcm []float32, data []byte) (int, error) {
	if enc.p == nil {
		return 0, errEncUninitialized
	}
	if len(pcm) == 0 {
		return 0, fmt.Errorf("opus: no data supplied")
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("opus: no target buffer")
	}
	if len(pcm)%enc.channels != 0 {
		return 0, fmt.Errorf("opus: input buffer length must be multiple of channels")
	}
	samples := len(pcm) / enc.channels
	n := int(C.opus_encode_float(
		enc.p,
		(*C.float)(&pcm[0]),
		C.int(samples),
		(*C.uchar)(&data[0]),
		C.opus_int32(cap(data))))
	if n < 0 {
		return 0, Error(n)
	}
	return n, nil
}

// SetDTX configures the encoder's use of discontinuous transmission (DTX).
func (enc *Encoder) SetDTX(dtx bool) error {
	i := 0
	if dtx {
		i = 1
	}
	res := C.bridge_encoder_set_dtx(enc.p, C.int(i))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// DTX reports whether this encoder is configured to use discontinuous
// transmission (DTX).
func (enc *Encoder) DTX() (bool, error) {
	var dtx C.int
	res := C.bridge_encoder_get_dtx(enc.p, &dtx)
	if res != C.OPUS_OK {
		return false, Error(res)
	}
	return dtx != 0, nil
}

func (enc *Encoder) SetVBR(vbr bool) error {
	i := 0
	if vbr {
		i = 1
	}
	res := C.bridge_encoder_set_vbr(enc.p, C.int(i))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

func (enc *Encoder) VBR() (bool, error) {
	var vbr C.int
	res := C.bridge_encoder_get_vbr(enc.p, &vbr)
	if res != C.OPUS_OK {
		return false, Error(res)
	}
	return vbr != 0, nil
}

func (enc *Encoder) SetVBRConstraint(vbrConstraint bool) error {
	i := 0
	if vbrConstraint {
		i = 1
	}
	res := C.bridge_encoder_set_vbr_constraint(enc.p, C.int(i))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

func (enc *Encoder) VBRConstraint() (bool, error) {
	var vbrConstraint C.int
	res := C.bridge_encoder_get_vbr_constraint(enc.p, &vbrConstraint)
	if res != C.OPUS_OK {
		return false, Error(res)
	}
	return vbrConstraint != 0, nil
}

func (enc *Encoder) SetSignalType(signal bool) error {
	i := 0
	if signal {
		i = 1
	}
	res := C.bridge_encoder_set_signal(enc.p, C.int(i))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

func (enc *Encoder) SignalType() (bool, error) {
	var signal C.int
	res := C.bridge_encoder_get_signal(enc.p, &signal)
	if res != C.OPUS_OK {
		return false, Error(res)
	}
	return signal != 0, nil
}

// SampleRate returns the encoder sample rate in Hz.
func (enc *Encoder) SampleRate() (int, error) {
	var sr C.opus_int32
	res := C.bridge_encoder_get_sample_rate(enc.p, &sr)
	if res != C.OPUS_OK {
		return 0, Error(res)
	}
	return int(sr), nil
}

// SetBitrate sets the bitrate of the Encoder
func (enc *Encoder) SetBitrate(bitrate int) error {
	res := C.bridge_encoder_set_bitrate(enc.p, C.opus_int32(bitrate))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// SetBitrateToAuto will allow the encoder to automatically set the bitrate
func (enc *Encoder) SetBitrateToAuto() error {
	res := C.bridge_encoder_set_bitrate(enc.p, C.opus_int32(C.OPUS_AUTO))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// SetBitrateToMax causes the encoder to use as much rate as it can. This can be
// useful for controlling the rate by adjusting the output buffer size.
func (enc *Encoder) SetBitrateToMax() error {
	res := C.bridge_encoder_set_bitrate(enc.p, C.opus_int32(C.OPUS_BITRATE_MAX))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// Bitrate returns the bitrate of the Encoder
func (enc *Encoder) Bitrate() (int, error) {
	var bitrate C.int
	res := C.bridge_encoder_get_bitrate(enc.p, &bitrate)
	if res != C.OPUS_OK {
		return 0, Error(res)
	}
	return int(bitrate), nil
}

// SetComplexity sets the encoder's computational complexity
func (enc *Encoder) SetComplexity(complexity int) error {
	res := C.bridge_encoder_set_complexity(enc.p, C.int(complexity))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// Complexity returns the computational complexity used by the encoder
func (enc *Encoder) Complexity() (int, error) {
	var complexity C.opus_int32
	res := C.bridge_encoder_get_complexity(enc.p, &complexity)
	if res != C.OPUS_OK {
		return 0, Error(res)
	}
	return int(complexity), nil
}

func (enc *Encoder) SetBandwidth(bandwidth Bandwidth) error {
	res := C.bridge_encoder_set_bandwidth(enc.p, C.int(bandwidth))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

func (enc *Encoder) Bandwidth() (Bandwidth, error) {
	var bandwidth C.int
	res := C.bridge_encoder_get_bandwidth(enc.p, &bandwidth)
	if res != C.OPUS_OK {
		return 0, Error(res)
	}
	return Bandwidth(bandwidth), nil
}

// SetMaxBandwidth configures the maximum bandpass that the encoder will select
// automatically
func (enc *Encoder) SetMaxBandwidth(maxBw Bandwidth) error {
	res := C.bridge_encoder_set_max_bandwidth(enc.p, C.int(maxBw))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// MaxBandwidth gets the encoder's configured maximum allowed bandpass.
func (enc *Encoder) MaxBandwidth() (Bandwidth, error) {
	var maxBw C.int
	res := C.bridge_encoder_get_max_bandwidth(enc.p, &maxBw)
	if res != C.OPUS_OK {
		return 0, Error(res)
	}
	return Bandwidth(maxBw), nil
}

// SetInBandFEC configures the encoder's use of inband forward error
// correction (FEC)
func (enc *Encoder) SetInBandFEC(fec bool) error {
	i := 0
	if fec {
		i = 1
	}
	res := C.bridge_encoder_set_inband_fec(enc.p, C.int(i))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// InBandFEC gets the encoder's configured inband forward error correction (FEC)
func (enc *Encoder) InBandFEC() (bool, error) {
	var fec C.int
	res := C.bridge_encoder_get_inband_fec(enc.p, &fec)
	if res != C.OPUS_OK {
		return false, Error(res)
	}
	return fec != 0, nil
}

// SetPacketLossPerc configures the encoder's expected packet loss percentage.
func (enc *Encoder) SetPacketLossPerc(lossPerc int) error {
	res := C.bridge_encoder_set_packet_loss_perc(enc.p, C.int(lossPerc))
	if res != C.OPUS_OK {
		return Error(res)
	}
	return nil
}

// PacketLossPerc gets the encoder's configured packet loss percentage.
func (enc *Encoder) PacketLossPerc() (int, error) {
	var lossPerc C.int
	res := C.bridge_encoder_get_packet_loss_perc(enc.p, &lossPerc)
	if res != C.OPUS_OK {
		return 0, Error(res)
	}
	return int(lossPerc), nil
}
