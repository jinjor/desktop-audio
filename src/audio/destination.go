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
destNoteFilterQ0V note_filter_q_0v
destNoteFilterGain note_filter_gain
destNoteFilterGain0V note_filter_gain_0v
destFilterFreq filter_freq
destFilterQ filter_q
destFilterQ0V filter_q_0v
destFilterGain filter_gain
destFilterGain0V filter_gain_0v
destLfo0Freq lfo0_freq
destLfo1Freq lfo1_freq
destLfo2Freq lfo2_freq
destLfo0Amount lfo0_amount
destLfo0Amount0V lfo0_amount_0v
destLfo1Amount lfo1_amount
destLfo1Amount0V lfo1_amount_0v
destLfo2Amount lfo2_amount
destLfo2Amount0V lfo2_amount_0v

EOF
*/

var destOscVolume = [2]int{destOsc0Volume, destOsc1Volume}
var destLfoFreq = [3]int{destLfo0Freq, destLfo1Freq, destLfo2Freq}
var destLfoAmount = [3]int{destLfo0Amount, destLfo1Amount, destLfo2Amount}
var destLfoAmount0V = [3]int{destLfo0Amount0V, destLfo1Amount0V, destLfo2Amount0V}
