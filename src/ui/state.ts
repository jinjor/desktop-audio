import * as d from "./decoder";
import { ipcRenderer } from "electron";
import { defaultLFO, lfoDecoder, lfoDestinations } from "./lfo";
import { oscDecoder } from "./osc";
import { adsrDecoder } from "./adsr";
import { defaultEnvelope, envelopeDecoder } from "./envelope";
import { noteFilterDecoder } from "./note-filter";
import { filterDecoder } from "./filter";
import { formantDecoder } from "./formant";
import { echoDecoder } from "./echo";
import {
  presetMetaDecoder,
  PresetAction,
  presetReducer,
  PresetState,
  initialPresetState,
  setPresetList,
} from "./preset";
import { ReducerWithEffect, ScheduleFn } from "./react-util";

const paramsDecoder = d.object({
  poly: d.string(),
  glideTime: d.number(),
  velSense: d.number(),
  oscs: d.array(oscDecoder),
  adsr: adsrDecoder,
  lfos: d.array(lfoDecoder),
  envelopes: d.array(envelopeDecoder),
  noteFilter: noteFilterDecoder,
  filter: filterDecoder,
  formant: formantDecoder,
  echo: echoDecoder,
});
const allParamsDecoder = d.object({
  name: d.optional(d.string(), null),
  params: paramsDecoder,
});
const statusDecoder = d.object({
  polyphony: d.number(),
  processTime: d.number(),
});
const presetListDecoder = d.object({
  items: d.array(presetMetaDecoder),
});
type Params = d.TypeOf<typeof paramsDecoder>;
type Status = d.TypeOf<typeof statusDecoder>;
export type State = {
  preset: PresetState;
  name: string | null;
  params: Params;
  status: Status;
};
export const initialState: State = {
  preset: initialPresetState,
  name: null,
  params: {
    poly: "mono",
    glideTime: 100,
    velSense: 0,
    oscs: [
      {
        enabled: true,
        kind: "sine",
        octave: 0,
        coarse: 0,
        fine: 0,
        level: 1.0,
      },
      {
        enabled: false,
        kind: "sine",
        octave: 0,
        coarse: 0,
        fine: 0,
        level: 1.0,
      },
    ],
    adsr: {
      attack: 0,
      decay: 100,
      sustain: 0.7,
      release: 100,
    },
    noteFilter: {
      enabled: false,
      targetOsc: "all",
      kind: "none",
      octave: 0,
      coarse: 0,
      q: 0,
      gain: 0,
    },
    filter: {
      enabled: false,
      targetOsc: "all",
      kind: "none",
      freq: 1000,
      q: 0,
      gain: 0,
    },
    formant: {
      enabled: false,
      kind: "a",
      tone: 1000,
      q: 0,
    },
    lfos: [defaultLFO, defaultLFO, defaultLFO],
    envelopes: [defaultEnvelope, defaultEnvelope, defaultEnvelope],
    echo: {
      enabled: false,
      delay: 100,
      feedbackGain: 0,
      mix: 0,
    },
  },
  status: {
    polyphony: 0,
    processTime: 0,
  },
};

export type Action =
  | { type: "receivedCommand"; command: string[] }
  | { type: "paramsAction"; value: ParamsAction }
  | { type: "presetAction"; value: PresetAction }
  | { type: "loadPreset"; value: string }
  | { type: "savePreset"; value: string }
  | { type: "removePreset"; value: string };
export type ParamsAction =
  | { type: "changedPoly"; value: string }
  | { type: "changedGlideTime"; value: number }
  | { type: "changedVelSense"; value: number }
  | { type: "changedOscEnabled"; index: number; value: boolean }
  | { type: "changedOscKind"; index: number; value: string }
  | { type: "changedOscOctave"; index: number; value: number }
  | { type: "changedOscCoarse"; index: number; value: number }
  | { type: "changedOscFine"; index: number; value: number }
  | { type: "changedOscLevel"; index: number; value: number }
  | { type: "changedAdsrAttack"; value: number }
  | { type: "changedAdsrDecay"; value: number }
  | { type: "changedAdsrSustain"; value: number }
  | { type: "changedAdsrRelease"; value: number }
  | { type: "changedLFOEnabled"; index: number; value: boolean }
  | { type: "changedLFODestination"; index: number; value: string }
  | { type: "changedLFOWave"; index: number; value: string }
  | { type: "changedLFOFreq"; index: number; value: number }
  | { type: "changedLFOAmount"; index: number; value: number }
  | { type: "changedEnvelopeEnabled"; index: number; value: boolean }
  | { type: "changedEnvelopeDestination"; index: number; value: string }
  | { type: "changedEnvelopeDelay"; index: number; value: number }
  | { type: "changedEnvelopeAttack"; index: number; value: number }
  | { type: "changedEnvelopeAmount"; index: number; value: number }
  | { type: "changedNoteFilterEnabled"; value: boolean }
  | { type: "changedNoteFilterTargetOsc"; value: string }
  | { type: "changedNoteFilterKind"; value: string }
  | { type: "changedNoteFilterOctave"; value: number }
  | { type: "changedNoteFilterCoarse"; value: number }
  | { type: "changedNoteFilterQ"; value: number }
  | { type: "changedNoteFilterGain"; value: number }
  | { type: "changedFilterEnabled"; value: boolean }
  | { type: "changedFilterTargetOsc"; value: string }
  | { type: "changedFilterKind"; value: string }
  | { type: "changedFilterFreq"; value: number }
  | { type: "changedFilterQ"; value: number }
  | { type: "changedFilterGain"; value: number }
  | { type: "changedFormantEnabled"; value: boolean }
  | { type: "changedFormantKind"; value: string }
  | { type: "changedFormantTone"; value: number }
  | { type: "changedFormantQ"; value: number }
  | { type: "changedEchoEnabled"; value: boolean }
  | { type: "changedEchoDelay"; value: number }
  | { type: "changedEchoFeedbackGain"; value: number }
  | { type: "changedEchoMix"; value: number };

const setItem = <T>(array: T[], index: number, updates: Partial<T>): T[] => {
  return array.map((item, i) => (i === index ? { ...item, ...updates } : item));
};

export const reducer: ReducerWithEffect<State, Action> = (
  state: State,
  action: Action,
  schedule: ScheduleFn<Action>
) => {
  switch (action.type) {
    case "receivedCommand": {
      const { command } = action;
      if (command[0] === "preset_list") {
        const obj = JSON.parse(command[1]);
        console.log(obj);
        const { items } = presetListDecoder.run(obj);
        return { ...state, preset: setPresetList(state.preset, items) };
      }
      if (command[0] === "all_params") {
        const obj = JSON.parse(command[1]);
        console.log(obj);
        const { name, params } = allParamsDecoder.run(obj);
        return { ...state, name, params };
      }
      if (command[0] === "status") {
        const obj = JSON.parse(command[1]);
        return { ...state, status: statusDecoder.run(obj) };
      } else {
        return state;
      }
    }
    case "paramsAction": {
      const { value } = action;
      return { ...state, params: paramsReducer(state.params, value) };
    }
    case "presetAction": {
      console.log(action);
      const { value } = action;
      const preset = presetReducer(state.preset, value, schedule);
      return { ...state, preset };
    }
    case "loadPreset": {
      const { value } = action;
      ipcRenderer.send("audio", ["preset", "load", value]);
      return state;
    }
    case "savePreset": {
      const { value } = action;
      ipcRenderer.send("audio", ["preset", "save_as", value]);
      return state;
    }
    case "removePreset": {
      const { value } = action;
      ipcRenderer.send("audio", ["preset", "remove", value]);
      return state;
    }
  }
};
const paramsReducer = (state: Params, action: ParamsAction): Params => {
  switch (action.type) {
    case "changedPoly": {
      const { value } = action;
      ipcRenderer.send("audio", [value]);
      return { ...state, poly: value };
    }
    case "changedGlideTime": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "glide_time", value]);
      return { ...state, glideTime: value };
    }
    case "changedVelSense": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "vel_sense", value]);
      return { ...state, velSense: value };
    }
    case "changedOscEnabled": {
      const { index, value } = action;
      ipcRenderer.send("audio", [
        "set",
        "osc",
        String(index),
        "enabled",
        String(value),
      ]);
      return {
        ...state,
        oscs: setItem(state.oscs, index, { enabled: value }),
      };
    }
    case "changedOscKind": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "osc", String(index), "kind", value]);
      return { ...state, oscs: setItem(state.oscs, index, { kind: value }) };
    }
    case "changedOscOctave": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "osc", String(index), "octave", value]);
      return {
        ...state,
        oscs: setItem(state.oscs, index, { octave: value }),
      };
    }
    case "changedOscCoarse": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "osc", String(index), "coarse", value]);
      return {
        ...state,
        oscs: setItem(state.oscs, index, { coarse: value }),
      };
    }
    case "changedOscFine": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "osc", String(index), "fine", value]);
      return { ...state, oscs: setItem(state.oscs, index, { fine: value }) };
    }
    case "changedOscLevel": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "osc", String(index), "level", value]);
      return { ...state, oscs: setItem(state.oscs, index, { level: value }) };
    }
    case "changedAdsrAttack": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "adsr", "attack", value]);
      return { ...state, adsr: { ...state.adsr, attack: value } };
    }
    case "changedAdsrDecay": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "adsr", "decay", value]);
      return { ...state, adsr: { ...state.adsr, decay: value } };
    }
    case "changedAdsrSustain": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "adsr", "sustain", value]);
      return { ...state, adsr: { ...state.adsr, sustain: value } };
    }
    case "changedAdsrRelease": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "adsr", "release", value]);
      return { ...state, adsr: { ...state.adsr, release: value } };
    }
    case "changedLFOEnabled": {
      const { index, value } = action;
      ipcRenderer.send("audio", [
        "set",
        "lfo",
        String(index),
        "enabled",
        String(value),
      ]);
      return {
        ...state,
        lfos: setItem(state.lfos, index, { enabled: value }),
      };
    }
    case "changedLFODestination": {
      const { index, value } = action;
      const { defaultFreq, defaultAmount } = lfoDestinations.get(value)!;
      ipcRenderer.send("audio", [
        "set",
        "lfo",
        String(index),
        "destination",
        value,
      ]);
      ipcRenderer.send("audio", [
        "set",
        "lfo",
        String(index),
        "freq",
        defaultFreq,
      ]);
      ipcRenderer.send("audio", [
        "set",
        "lfo",
        String(index),
        "amount",
        defaultAmount,
      ]);
      return {
        ...state,
        lfos: setItem(state.lfos, index, {
          destination: value,
          freq: defaultFreq,
          amount: defaultAmount,
        }),
      };
    }
    case "changedLFOWave": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "lfo", String(index), "wave", value]);
      return {
        ...state,
        lfos: setItem(state.lfos, index, { wave: value }),
      };
    }
    case "changedLFOFreq": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "lfo", String(index), "freq", value]);
      return {
        ...state,
        lfos: setItem(state.lfos, index, { freq: value }),
      };
    }
    case "changedLFOAmount": {
      const { index, value } = action;
      ipcRenderer.send("audio", ["set", "lfo", String(index), "amount", value]);
      return {
        ...state,
        lfos: setItem(state.lfos, index, { amount: value }),
      };
    }
    case "changedEnvelopeEnabled": {
      const { index, value } = action;
      ipcRenderer.send("audio", [
        "set",
        "envelope",
        String(index),
        "enabled",
        String(value),
      ]);
      return {
        ...state,
        envelopes: setItem(state.envelopes, index, { enabled: value }),
      };
    }
    case "changedEnvelopeDestination": {
      const { index, value } = action;
      ipcRenderer.send("audio", [
        "set",
        "envelope",
        String(index),
        "destination",
        value,
      ]);
      return {
        ...state,
        envelopes: setItem(state.envelopes, index, {
          destination: value,
        }),
      };
    }
    case "changedEnvelopeDelay": {
      const { index, value } = action;
      ipcRenderer.send("audio", [
        "set",
        "envelope",
        String(index),
        "delay",
        value,
      ]);
      return {
        ...state,
        envelopes: setItem(state.envelopes, index, {
          delay: value,
        }),
      };
    }
    case "changedEnvelopeAttack": {
      const { index, value } = action;
      ipcRenderer.send("audio", [
        "set",
        "envelope",
        String(index),
        "attack",
        value,
      ]);
      return {
        ...state,
        envelopes: setItem(state.envelopes, index, {
          attack: value,
        }),
      };
    }
    case "changedEnvelopeAmount": {
      const { index, value } = action;
      ipcRenderer.send("audio", [
        "set",
        "envelope",
        String(index),
        "amount",
        value,
      ]);
      return {
        ...state,
        envelopes: setItem(state.envelopes, index, {
          amount: value,
        }),
      };
    }
    case "changedNoteFilterEnabled": {
      const { value } = action;
      ipcRenderer.send("audio", [
        "set",
        "note_filter",
        "enabled",
        String(value),
      ]);
      return {
        ...state,
        noteFilter: { ...state.noteFilter, enabled: value },
      };
    }
    case "changedNoteFilterTargetOsc": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "note_filter", "target_osc", value]);
      return {
        ...state,
        noteFilter: { ...state.noteFilter, targetOsc: value },
      };
    }
    case "changedNoteFilterKind": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "note_filter", "kind", value]);
      return {
        ...state,
        noteFilter: { ...state.noteFilter, kind: value },
      };
    }
    case "changedNoteFilterOctave": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "note_filter", "octave", value]);
      return {
        ...state,
        noteFilter: { ...state.noteFilter, octave: value },
      };
    }
    case "changedNoteFilterCoarse": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "note_filter", "coarse", value]);
      return {
        ...state,
        noteFilter: { ...state.noteFilter, coarse: value },
      };
    }
    case "changedNoteFilterQ": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "note_filter", "q", value]);
      return {
        ...state,
        noteFilter: { ...state.noteFilter, q: value },
      };
    }
    case "changedNoteFilterGain": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "note_filter", "gain", value]);
      return {
        ...state,
        noteFilter: { ...state.noteFilter, gain: value },
      };
    }
    case "changedFilterEnabled": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "filter", "enabled", String(value)]);
      return {
        ...state,
        filter: { ...state.filter, enabled: value },
      };
    }
    case "changedFilterTargetOsc": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "filter", "target_osc", value]);
      return {
        ...state,
        filter: { ...state.filter, targetOsc: value },
      };
    }
    case "changedFilterKind": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "filter", "kind", value]);
      return {
        ...state,
        filter: { ...state.filter, kind: value },
      };
    }
    case "changedFilterFreq": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "filter", "freq", value]);
      return {
        ...state,
        filter: { ...state.filter, freq: value },
      };
    }
    case "changedFilterQ": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "filter", "q", value]);
      return {
        ...state,
        filter: { ...state.filter, q: value },
      };
    }
    case "changedFilterGain": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "filter", "gain", value]);
      return {
        ...state,
        filter: { ...state.filter, gain: value },
      };
    }
    case "changedFormantEnabled": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "formant", "enabled", String(value)]);
      return {
        ...state,
        formant: { ...state.formant, enabled: value },
      };
    }
    case "changedFormantKind": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "formant", "kind", value]);
      return {
        ...state,
        formant: { ...state.formant, kind: value },
      };
    }
    case "changedFormantTone": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "formant", "tone", value]);
      return {
        ...state,
        formant: { ...state.formant, tone: value },
      };
    }
    case "changedFormantQ": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "formant", "q", value]);
      return {
        ...state,
        formant: { ...state.formant, q: value },
      };
    }
    case "changedEchoEnabled": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "echo", "enabled", String(value)]);
      return {
        ...state,
        echo: { ...state.echo, enabled: value },
      };
    }
    case "changedEchoDelay": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "echo", "delay", value]);
      return {
        ...state,
        echo: { ...state.echo, delay: value },
      };
    }
    case "changedEchoFeedbackGain": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "echo", "feedbackGain", value]);
      return {
        ...state,
        echo: { ...state.echo, feedbackGain: value },
      };
    }
    case "changedEchoMix": {
      const { value } = action;
      ipcRenderer.send("audio", ["set", "echo", "mix", value]);
      return {
        ...state,
        echo: { ...state.echo, mix: value },
      };
    }
  }
};
