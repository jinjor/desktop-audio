package audio

// ----- Destination ----- //

//go:generate go run ../gen/main.go -- destination.gen.go
/*
generate-enum destination

destNone none
destOsc0Volume osc0_volume
destOsc1Volume osc1_volume
destVibrato vibrato
destTremolo tremolo
destFM fm
destPM pm
destAM am
destFreq freq
destNoteFilterFreq note_filter_freq
destNoteFilterQ note_filter_q
destNoteFilterGain note_filter_gain
destFilterFreq filter_freq
destFilterQ filter_q
destFilterGain filter_gain
destLfo0Freq lfo0_freq
destLfo1Freq lfo1_freq
destLfo2Freq lfo2_freq
destLfo0Amount lfo0_amount
destLfo1Amount lfo1_amount
destLfo2Amount lfo2_amount

EOF
*/

var destOscVolume = [2]int{destOsc0Volume, destOsc1Volume}
var destLfoFreq = [3]int{destLfo0Freq, destLfo1Freq, destLfo2Freq}
var destLfoAmount = [3]int{destLfo0Amount, destLfo1Amount, destLfo2Amount}
