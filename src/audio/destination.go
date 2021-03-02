package audio

// ----- Destination ----- //

//go:generate go run ../gen/main.go -- destination.gen.go
/*
generate-enum destination

destNone none
destVibrato vibrato
destTremolo tremolo
destFM fm
destPM pm
destAM am
destFreq freq
destFilterFreq filter_freq
destFilterQ filter_q
destFilterQ0V filter_q_0v
destFilterGain filter_gain
destFilterGain0V filter_gain_0v
destLfo0Freq lfo0_freq
destLfo1Freq lfo1_freq
destLfo2Freq lfo2_freq
destLfo0Amount lfo0_amount
destLfo1Amount lfo1_amount
destLfo2Amount lfo2_amount

EOF
*/

var destLfoFreq = [3]int{destLfo0Freq, destLfo1Freq, destLfo2Freq}
var destLfoAmount = [3]int{destLfo0Amount, destLfo1Amount, destLfo2Amount}
