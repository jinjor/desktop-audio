import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useEffect, useRef, useCallback, useReducer } from "react";
import { Notes } from "./note";
import { LabeledKnob } from "./knob";
import { Radio } from "./radio";
import { Select } from "./select";
import * as d from "./decoder";
import { WaveSelect } from "./wave-select";

const EditGroup = (o: {
  enabled?: boolean;
  label: string;
  children: any;
  canBypass?: boolean;
  onChangeEnabled?: (value: boolean) => void;
}) => {
  const enabled = o.enabled ?? true;
  const canBypass = o.canBypass ?? false;
  const onClick = () => {
    if (canBypass) {
      o.onChangeEnabled?.(!enabled);
    }
  };
  return (
    <div
      style={{
        display: "flex",
        flexFlow: "column",
      }}
    >
      <div
        style={{
          display: "flex",
          borderBottom: "solid 1px #aaa",
          whiteSpace: "nowrap",
          alignItems: "center",
          columnGap: "4px",
        }}
        onClick={onClick}
      >
        {o.canBypass ? (
          <div
            style={{
              width: "8px",
              height: "8px",
              backgroundColor: enabled ? "rgb(153,119,255)" : "#222",
              marginTop: "-1px",
            }}
          ></div>
        ) : null}
        <label>{o.label}</label>
      </div>
      <div
        style={{
          padding: "5px 0",
          ...(enabled ? {} : { opacity: 0.2, pointerEvents: "none" }),
        }}
      >
        {o.children}
      </div>
    </div>
  );
};

const Poly = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedPoly", value });
    return (
      <Radio list={["mono", "poly"]} value={o.value} onChange={onChange} />
    );
  }
);
const GlideTime = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedGlideTime", value });
    const min = 1;
    const max = 400;
    const steps = max - min + 1;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        from={0}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Glide"
      />
    );
  }
);
const VelSense = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedVelSense", value });
    const min = 0;
    const max = 1;
    const steps = 400;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        from={0}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Sense"
      />
    );
  }
);
type OSC = {
  enabled: boolean;
  kind: string;
  octave: number;
  coarse: number;
  fine: number;
  level: number;
};
const OSCGroup = React.memo(
  (o: {
    index: number;
    value: OSC;
    dispatch: React.Dispatch<ParamsAction>;
  }) => {
    const onChangeEnabled = useCallbackWithIndex(
      o.index,
      (index: number, value: boolean) =>
        o.dispatch({ type: "changedOscEnabled", index, value })
    );
    const onChangeKind = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedOscKind", index, value })
    );
    const onChangeOctave = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscOctave", index, value })
    );
    const onChangeCoarse = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscCoarse", index, value })
    );
    const onChangeFine = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscFine", index, value })
    );
    const onChangeLevel = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscLevel", index, value })
    );
    return (
      <EditGroup
        label={`OSC ${o.index + 1}`}
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <div style={{ textAlign: "center" }}>
            <OscKind value={o.value.kind} onChange={onChangeKind} />
          </div>
          <OSCOctave value={o.value.octave} onChange={onChangeOctave} />
          <OSCCoarse value={o.value.coarse} onChange={onChangeCoarse} />
          <OSCFine value={o.value.fine} onChange={onChangeFine} />
          <OSCLevel value={o.value.level} onChange={onChangeLevel} />
        </div>
      </EditGroup>
    );
  }
);

const OscKind = React.memo(
  (o: { onChange: (value: string) => void; value: string }) => {
    return (
      <WaveSelect
        list={[
          "sine",
          "triangle",
          // "square",
          "square-wt",
          "pulse",
          // "saw",
          "saw-wt",
          "noise",
        ]}
        value={o.value}
        columns={3}
        onChange={o.onChange}
      />
    );
  }
);
const OSCOctave = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = -2;
    const max = 2;
    const steps = max - min + 1;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        from={0}
        exponential={true}
        value={o.value}
        onChange={o.onChange}
        label="Octave"
      />
    );
  }
);
const OSCCoarse = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = -12;
    const max = 12;
    const steps = max - min + 1;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        exponential={true}
        value={o.value}
        from={0}
        onChange={o.onChange}
        label="Coarse"
      />
    );
  }
);
const OSCFine = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = -100;
    const max = 100;
    const steps = max - min + 1;
    const onChange = (value: number) => o.onChange(Math.round(value));
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        exponential={true}
        value={o.value}
        from={0}
        onChange={onChange}
        label="Fine"
      />
    );
  }
);
const OSCLevel = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = 0;
    const max = 1;
    const steps = 400;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        exponential={false}
        value={o.value}
        from={0}
        onChange={o.onChange}
        label="Level"
      />
    );
  }
);
const Attack = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrAttack", value });
    return (
      <LabeledKnob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Attack"
      />
    );
  }
);
const Decay = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrDecay", value });
    return (
      <LabeledKnob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Decay"
      />
    );
  }
);
const Sustain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrSustain", value });
    return (
      <LabeledKnob
        min={0}
        max={1}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Sustain"
      />
    );
  }
);
const Release = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrRelease", value });
    return (
      <LabeledKnob
        min={0}
        max={800}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Release"
      />
    );
  }
);

const NoteFilterKind = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedNoteFilterKind", value });
    return (
      <Select
        list={[
          "none",
          "lowpass",
          "highpass",
          "bandpass-1",
          "bandpass-2",
          "peaking",
          "lowshelf",
          "highshelf",
        ]}
        value={o.value}
        onChange={onChange}
      />
    );
  }
);
const NoteFilterBaseOsc = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const list = ["OSC 1", "OSC 2"];
    const onChange = (value: string) =>
      o.dispatch({
        type: "changedNoteFilterBaseOsc",
        value: list.indexOf(value),
      });
    return <Select list={list} value={list[o.value]} onChange={onChange} />;
  }
);
const NoteFilterOctave = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({
        type: "changedNoteFilterOctave",
        value,
      });
    const min = -2;
    const max = 2;
    const steps = max - min + 1;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        from={0}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Octave"
      />
    );
  }
);
const NoteFilterCoarse = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({
        type: "changedNoteFilterCoarse",
        value,
      });
    const min = -12;
    const max = 12;
    const steps = max - min + 1;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        exponential={true}
        value={o.value}
        from={0}
        onChange={onChange}
        label="Coarse"
      />
    );
  }
);
const NoteFilterQ = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedNoteFilterQ", value });
    return (
      <LabeledKnob
        min={0}
        max={20}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Q"
      />
    );
  }
);
const NoteFilterGain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedNoteFilterGain", value });
    return (
      <LabeledKnob
        min={-40}
        max={40}
        steps={400}
        exponential={false}
        value={o.value}
        from={0}
        onChange={onChange}
        label="Gain"
      />
    );
  }
);

const FilterKind = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedFilterKind", value });
    return (
      <Radio
        list={[
          "none",
          "lowpass-fir",
          "highpass-fir",
          "lowpass",
          "highpass",
          "bandpass-1",
          "bandpass-2",
          "notch",
          "peaking",
          "lowshelf",
          "highshelf",
        ]}
        value={o.value}
        onChange={onChange}
      />
    );
  }
);
const FilterFreq = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFilterFreq", value });
    return (
      <LabeledKnob
        min={30}
        max={20000}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Freq"
      />
    );
  }
);
const FilterQ = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFilterQ", value });
    return (
      <LabeledKnob
        min={0}
        max={20}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Q"
      />
    );
  }
);
const FilterGain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFilterGain", value });
    return (
      <LabeledKnob
        min={-40}
        max={40}
        steps={400}
        exponential={false}
        value={o.value}
        from={0}
        onChange={onChange}
        label="Gain"
      />
    );
  }
);

const FormantKind = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedFormantKind", value });
    return (
      <Radio
        list={["a", "e", "i", "o", "u"]}
        value={o.value}
        onChange={onChange}
      />
    );
  }
);
const FormantTone = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFormantTone", value });
    return (
      <LabeledKnob
        min={0.5}
        max={2}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Tone"
      />
    );
  }
);
const FormantQ = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFormantQ", value });
    return (
      <LabeledKnob
        min={0}
        max={50}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Q"
      />
    );
  }
);

type LFODestination = {
  freqType: string;
  minFreq: number;
  maxFreq: number;
  defaultFreq: number;
  minAmount: number;
  maxAmount: number;
  defaultAmount: number;
  fromAmount: number;
};
const lfoDestinations = new Map<string, LFODestination>();
lfoDestinations.set("none", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 0,
  defaultFreq: 0,
  minAmount: 0,
  maxAmount: 0,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("tremolo", {
  freqType: "absolute",
  minFreq: 0.1,
  maxFreq: 100,
  defaultFreq: 0.1,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("vibrato", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 40,
  defaultFreq: 10,
  minAmount: 0,
  maxAmount: 200,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("fm", {
  freqType: "ratio",
  minFreq: 0.1,
  maxFreq: 10,
  defaultFreq: 3,
  minAmount: -2400,
  maxAmount: 2400,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("pm", {
  freqType: "ratio",
  minFreq: 0.1,
  maxFreq: 10,
  defaultFreq: 3,
  minAmount: 0,
  maxAmount: 1.57,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("am", {
  freqType: "ratio",
  minFreq: 0.1,
  maxFreq: 10,
  defaultFreq: 3,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("filter_freq", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 40,
  defaultFreq: 10,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});
type LFO = {
  enabled: boolean;
  destination: string;
  wave: string;
  freq: number;
  amount: number;
};
const defaultLFO = {
  enabled: false,
  destination: "none",
  wave: "sine",
  freq: 0,
  amount: 0,
};
const useCallbackWithIndex = <T,>(
  index: number,
  f: (index: number, value: T) => void
) => {
  return useCallback((value: T) => f(index, value), [index, f]);
};
const LFOGroup = React.memo(
  (o: {
    index: number;
    value: LFO;
    dispatch: React.Dispatch<ParamsAction>;
  }) => {
    const list = [...lfoDestinations.keys()];
    const destination = lfoDestinations.get(o.value.destination)!;
    const onChangeEnabled = useCallbackWithIndex(
      o.index,
      (index: number, value: boolean) =>
        o.dispatch({ type: "changedLFOEnabled", index, value })
    );
    const onChangeDestination = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedLFODestination", index, value })
    );
    const onChangeWave = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedLFOWave", index, value })
    );
    const onChangeFreq = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedLFOFreq", index, value })
    );
    const onChangeAmount = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedLFOAmount", index, value })
    );
    return (
      <EditGroup
        label={`LFO ${o.index + 1}`}
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <Select
            list={list}
            value={o.value.destination}
            onChange={onChangeDestination}
          />
          <div style={{ textAlign: "center" }}>
            <LFOWave value={o.value.wave} onChange={onChangeWave} />
          </div>
          <LFOFreq
            min={destination.minFreq}
            max={destination.maxFreq}
            value={o.value.freq}
            onChange={onChangeFreq}
          />
          <LFOAmount
            min={destination.minAmount}
            max={destination.maxAmount}
            value={o.value.amount}
            from={destination.fromAmount}
            onChange={onChangeAmount}
          />
        </div>
      </EditGroup>
    );
  }
);
const LFOWave = React.memo(
  (o: { value: string; onChange: (value: string) => void }) => {
    return (
      <WaveSelect
        list={[
          "sine",
          "triangle",
          // "square",
          "square-wt",
          "pulse",
          // "saw",
          "saw-wt",
          "saw-rev",
        ]}
        value={o.value}
        columns={3}
        onChange={o.onChange}
      />
    );
  }
);
const LFOFreq = React.memo(
  (o: {
    min: number;
    max: number;
    value: number;
    onChange: (value: number) => void;
  }) => {
    return (
      <LabeledKnob
        min={o.min}
        max={o.max}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Freq"
      />
    );
  }
);
const LFOAmount = React.memo(
  (o: {
    min: number;
    max: number;
    value: number;
    from: number;
    onChange: (value: number) => void;
  }) => {
    return (
      <LabeledKnob
        min={o.min}
        max={o.max}
        steps={400}
        exponential={false}
        value={o.value}
        from={o.from}
        onChange={o.onChange}
        label="Amount"
      />
    );
  }
);

type EnvelopeDestination = {
  minDelay: number;
  maxDelay: number;
  defaultDelay: number;
};
const envelopeDestinations = new Map<string, EnvelopeDestination>();
envelopeDestinations.set("none", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("freq", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_freq", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_q", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_q_0v", {
  minDelay: 0,
  maxDelay: 1000,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_gain", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_gain_0v", {
  minDelay: 0,
  maxDelay: 1000,
  defaultDelay: 0,
});
for (let i = 0; i < 3; i++) {
  envelopeDestinations.set(`lfo${i}_amount`, {
    minDelay: 0,
    maxDelay: 1000,
    defaultDelay: 200,
  });
}
for (let i = 0; i < 3; i++) {
  envelopeDestinations.set(`lfo${i}_freq`, {
    minDelay: 0,
    maxDelay: 0,
    defaultDelay: 0,
  });
}
type Envelope = {
  enabled: boolean;
  destination: string;
  delay: number;
  attack: number;
  amount: number;
};
const defaultEnvelope: Envelope = {
  enabled: false,
  destination: "none",
  delay: 0,
  attack: 0,
  amount: 0,
};
const EnvelopeGroup = React.memo(
  (o: {
    index: number;
    value: Envelope;
    dispatch: React.Dispatch<ParamsAction>;
  }) => {
    const list = [...envelopeDestinations.keys()];
    const destination = envelopeDestinations.get(o.value.destination)!;
    const onChangeEnabled = useCallbackWithIndex(
      o.index,
      (index: number, value: boolean) =>
        o.dispatch({ type: "changedEnvelopeEnabled", index, value })
    );
    const onChangeDestination = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedEnvelopeDestination", index, value })
    );
    const onChangeDelay = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedEnvelopeDelay", index, value })
    );
    const onChangeAttack = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedEnvelopeAttack", index, value })
    );
    const onChangeAmount = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedEnvelopeAmount", index, value })
    );
    return (
      <EditGroup
        label={`Envelope ${o.index + 1}`}
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <Select
            list={list}
            value={o.value.destination}
            onChange={onChangeDestination}
          />
          <EnvelopeDelay
            min={destination.minDelay}
            max={destination.maxDelay}
            value={o.value.delay}
            onChange={onChangeDelay}
          />
          <EnvelopeAttack value={o.value.attack} onChange={onChangeAttack} />
          <EnvelopeAmount value={o.value.amount} onChange={onChangeAmount} />
        </div>
      </EditGroup>
    );
  }
);
const EnvelopeDelay = React.memo(
  (o: {
    min: number;
    max: number;
    value: number;
    onChange: (value: number) => void;
  }) => {
    return (
      <LabeledKnob
        min={o.min}
        max={o.max}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Delay"
      />
    );
  }
);
const EnvelopeAttack = React.memo(
  (o: { value: number; onChange: (value: number) => void }) => {
    return (
      <LabeledKnob
        min={0}
        max={1000}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Attack"
      />
    );
  }
);
const EnvelopeAmount = React.memo(
  (o: { value: number; onChange: (value: number) => void }) => {
    return (
      <LabeledKnob
        min={-1}
        max={1}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Amount"
      />
    );
  }
);
const EchoDelay = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedEchoDelay", value });
    return (
      <LabeledKnob
        min={10}
        max={800}
        steps={791}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Delay"
      />
    );
  }
);
const EchoFeedbackGain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedEchoFeedbackGain", value });
    return (
      <LabeledKnob
        min={0}
        max={0.5}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="FeedbackGain"
      />
    );
  }
);
const EchoMix = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedEchoMix", value });
    return (
      <LabeledKnob
        min={0}
        max={1.0}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Mix"
      />
    );
  }
);
const Canvas = (props: {
  listen: (canvas: HTMLCanvasElement) => () => void;
  [key: string]: any;
}) => {
  const { listen, ...canvasProps } = props;
  const el: React.MutableRefObject<HTMLCanvasElement | null> = useRef(null);
  useEffect(() => listen(el.current!), []);
  return <canvas {...canvasProps} ref={el}></canvas>;
};

const Spectrum = () => {
  const listen = (canvas: HTMLCanvasElement) => {
    const width = canvas.width;
    const height = canvas.height;
    const sampleRate = 48000;
    const maxFreq = 24000;
    const minFreq = 32;
    const fftData: number[] = [];
    const filterShapeData: number[] = [];
    const render = () => {
      const ctx = canvas.getContext("2d")!;
      ctx.fillStyle = "black";
      ctx.fillRect(0, 0, width, height);
      renderFrequencyShape(ctx, fftData, {
        color: "#66dd66",
        sampleRate,
        width,
        height,
        minFreq,
        maxFreq,
        minDb: -100,
        maxDb: -6,
      });
      renderFrequencyShape(ctx, filterShapeData, {
        color: "#66aadd",
        sampleRate,
        width,
        height,
        minFreq,
        maxFreq,
        minDb: -50,
        maxDb: 50,
      });
    };
    const callback = (_: any, command: string[]) => {
      if (command[0] === "fft") {
        for (let i = 1; i < command.length; i++) {
          fftData[i - 1] = parseFloat(command[i]);
        }
        render();
      } else if (command[0] === "filter-shape") {
        for (let i = 1; i < command.length; i++) {
          filterShapeData[i - 1] = parseFloat(command[i]);
        }
        render();
      }
    };
    ipcRenderer.on("audio", callback);
    return () => {
      ipcRenderer.off("audio", callback);
    };
  };
  return <Canvas width="256" height="100" listen={listen} />;
};

function renderFrequencyShape(
  ctx: CanvasRenderingContext2D,
  data: number[],
  o: {
    color: string;
    sampleRate: number;
    width: number;
    height: number;
    minFreq: number;
    maxFreq: number;
    minDb: number;
    maxDb: number;
  }
) {
  ctx.strokeStyle = o.color;
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(0, o.height);
  for (let i = 0; i < data.length; i++) {
    const value = data[i];
    const freq = (o.sampleRate / 2) * (i / data.length);
    const x =
      (Math.log(freq / o.minFreq) / Math.log(o.maxFreq / o.minFreq)) * o.width;
    const db = 20 * Math.log10(value);
    const y = (1 - (db - o.minDb) / (o.maxDb - o.minDb)) * o.height;
    ctx.lineTo(x, y);
  }
  ctx.stroke();
}

const setItem = <T,>(array: T[], index: number, updates: Partial<T>): T[] => {
  return array.map((item, i) => (i === index ? { ...item, ...updates } : item));
};

const paramsDecoder = d.object({
  poly: d.string(),
  glideTime: d.number(),
  velSense: d.number(),
  oscs: d.array(
    d.object({
      enabled: d.boolean,
      kind: d.string(),
      octave: d.number(),
      coarse: d.number(),
      fine: d.number(),
      level: d.number(),
    })
  ),
  adsr: d.object({
    attack: d.number(),
    decay: d.number(),
    sustain: d.number(),
    release: d.number(),
  }),
  lfos: d.array(
    d.object({
      enabled: d.boolean,
      destination: d.string(),
      wave: d.string(),
      freq: d.number(),
      amount: d.number(),
    })
  ),
  envelopes: d.array(
    d.object({
      enabled: d.boolean,
      destination: d.string(),
      delay: d.number(),
      attack: d.number(),
      amount: d.number(),
    })
  ),
  noteFilter: d.object({
    enabled: d.boolean,
    baseOsc: d.number(),
    kind: d.string(),
    octave: d.number(),
    coarse: d.number(),
    q: d.number(),
    gain: d.number(),
  }),
  filter: d.object({
    enabled: d.boolean,
    kind: d.string(),
    freq: d.number(),
    q: d.number(),
    gain: d.number(),
  }),
  formant: d.object({
    enabled: d.boolean,
    kind: d.string(),
    tone: d.number(),
    q: d.number(),
  }),
  echo: d.object({
    enabled: d.boolean,
    delay: d.number(),
    feedbackGain: d.number(),
    mix: d.number(),
  }),
});
const statusDecoder = d.object({
  polyphony: d.number(),
  processTime: d.number(),
});
const stateDecoder = d.object({
  params: paramsDecoder,
  status: statusDecoder,
});
type Params = d.TypeOf<typeof paramsDecoder>;
type State = d.TypeOf<typeof stateDecoder>;

type Action =
  | { type: "receivedCommand"; command: string[] }
  | { type: "paramsAction"; value: ParamsAction };
type ParamsAction =
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
  | { type: "changedNoteFilterBaseOsc"; value: number }
  | { type: "changedNoteFilterKind"; value: string }
  | { type: "changedNoteFilterOctave"; value: number }
  | { type: "changedNoteFilterCoarse"; value: number }
  | { type: "changedNoteFilterQ"; value: number }
  | { type: "changedNoteFilterGain"; value: number }
  | { type: "changedFilterEnabled"; value: boolean }
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

const App = () => {
  const initialState: State = {
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
        baseOsc: 0,
        kind: "none",
        octave: 0,
        coarse: 0,
        q: 0,
        gain: 0,
      },
      filter: {
        enabled: false,
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
        ipcRenderer.send("audio", [
          "set",
          "osc",
          String(index),
          "octave",
          value,
        ]);
        return {
          ...state,
          oscs: setItem(state.oscs, index, { octave: value }),
        };
      }
      case "changedOscCoarse": {
        const { index, value } = action;
        ipcRenderer.send("audio", [
          "set",
          "osc",
          String(index),
          "coarse",
          value,
        ]);
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
        ipcRenderer.send("audio", [
          "set",
          "osc",
          String(index),
          "level",
          value,
        ]);
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
        ipcRenderer.send("audio", [
          "set",
          "lfo",
          String(index),
          "amount",
          value,
        ]);
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
      case "changedNoteFilterBaseOsc": {
        const { value } = action;
        ipcRenderer.send("audio", ["set", "note_filter", "base_osc", value]);
        return {
          ...state,
          noteFilter: { ...state.noteFilter, baseOsc: value },
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
  const [state, dispatch] = useReducer(
    (state: State, action: Action): State => {
      switch (action.type) {
        case "receivedCommand": {
          const { command } = action;
          if (command[0] === "all_params") {
            const obj = JSON.parse(command[1]);
            console.log(obj);
            return { ...state, params: paramsDecoder.run(obj) };
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
      }
    },
    initialState
  );
  useEffect(() => {
    const callback = (_: any, command: string[]) => {
      dispatch({ type: "receivedCommand", command });
    };
    ipcRenderer.on("audio", callback);
    return () => {
      ipcRenderer.off("audio", callback);
    };
  }, []);
  const dispatchParam: React.Dispatch<ParamsAction> = (
    action: ParamsAction
  ) => {
    dispatch({
      type: "paramsAction",
      value: action,
    });
  };
  const onChangeNoteFilterEnabled = (value: boolean) => {
    dispatchParam({
      type: "changedNoteFilterEnabled",
      value,
    });
  };
  const onChangeFilterEnabled = (value: boolean) => {
    dispatchParam({
      type: "changedFilterEnabled",
      value,
    });
  };
  const onChangeFormantEnabled = (value: boolean) => {
    dispatchParam({
      type: "changedFormantEnabled",
      value,
    });
  };
  const onChangeEchoEnabled = (value: boolean) => {
    dispatchParam({
      type: "changedEchoEnabled",
      value,
    });
  };
  const p = state.params;
  const processTimeLimit = 0.0213; // TODO: get this from server
  return (
    <React.Fragment>
      <div style={{ display: "flex", gap: "20px", padding: "5px 10px" }}>
        <EditGroup label="NOTE">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <Poly value={p.poly} dispatch={dispatchParam} />
            <GlideTime value={p.glideTime} dispatch={dispatchParam} />
            <VelSense value={p.velSense} dispatch={dispatchParam} />
          </div>
        </EditGroup>
        <OSCGroup index={0} value={p.oscs[0]} dispatch={dispatchParam} />
        <OSCGroup index={1} value={p.oscs[1]} dispatch={dispatchParam} />
        <EditGroup label="Envelope">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <Attack value={p.adsr.attack} dispatch={dispatchParam} />
            <Decay value={p.adsr.decay} dispatch={dispatchParam} />
            <Sustain value={p.adsr.sustain} dispatch={dispatchParam} />
            <Release value={p.adsr.release} dispatch={dispatchParam} />
          </div>
        </EditGroup>
        <EditGroup
          label="Note Filter"
          enabled={p.noteFilter.enabled}
          canBypass={true}
          onChangeEnabled={onChangeNoteFilterEnabled}
        >
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <NoteFilterKind
              dispatch={dispatchParam}
              value={p.noteFilter.kind}
            />
            <NoteFilterBaseOsc
              dispatch={dispatchParam}
              value={p.noteFilter.baseOsc}
            />
            <NoteFilterOctave
              dispatch={dispatchParam}
              value={p.noteFilter.octave}
            />
            <NoteFilterCoarse
              dispatch={dispatchParam}
              value={p.noteFilter.coarse}
            />
            <NoteFilterQ dispatch={dispatchParam} value={p.noteFilter.q} />
            <NoteFilterGain
              dispatch={dispatchParam}
              value={p.noteFilter.gain}
            />
          </div>
        </EditGroup>
        <EditGroup
          label="Filter"
          enabled={p.filter.enabled}
          canBypass={true}
          onChangeEnabled={onChangeFilterEnabled}
        >
          <div style={{ display: "flex", gap: "12px" }}>
            <div>
              <FilterKind dispatch={dispatchParam} value={p.filter.kind} />
            </div>
            <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
              <FilterFreq dispatch={dispatchParam} value={p.filter.freq} />
              <FilterQ dispatch={dispatchParam} value={p.filter.q} />
              <FilterGain dispatch={dispatchParam} value={p.filter.gain} />
            </div>
          </div>
        </EditGroup>
        <EditGroup
          label="Formant"
          enabled={p.formant.enabled}
          canBypass={true}
          onChangeEnabled={onChangeFormantEnabled}
        >
          <div style={{ display: "flex", gap: "12px" }}>
            <div>
              <FormantKind dispatch={dispatchParam} value={p.formant.kind} />
            </div>
            <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
              <FormantTone dispatch={dispatchParam} value={p.formant.tone} />
              <FormantQ dispatch={dispatchParam} value={p.formant.q} />
            </div>
          </div>
        </EditGroup>
        <LFOGroup index={0} value={p.lfos[0]} dispatch={dispatchParam} />
        <LFOGroup index={1} value={p.lfos[1]} dispatch={dispatchParam} />
        <LFOGroup index={2} value={p.lfos[2]} dispatch={dispatchParam} />
        <EnvelopeGroup
          index={0}
          value={p.envelopes[0]}
          dispatch={dispatchParam}
        />
        <EnvelopeGroup
          index={1}
          value={p.envelopes[1]}
          dispatch={dispatchParam}
        />
        <EnvelopeGroup
          index={2}
          value={p.envelopes[2]}
          dispatch={dispatchParam}
        />
        <EditGroup
          label="Echo"
          enabled={p.echo.enabled}
          canBypass={true}
          onChangeEnabled={onChangeEchoEnabled}
        >
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <EchoDelay value={p.echo.delay} dispatch={dispatchParam} />
            <EchoFeedbackGain
              value={p.echo.feedbackGain}
              dispatch={dispatchParam}
            />
            <EchoMix value={p.echo.mix} dispatch={dispatchParam} />
          </div>
        </EditGroup>
      </div>
      <Notes />
      <Spectrum />
      <div>Polyphony: {state.status.polyphony}</div>
      <div>
        Process Time: {(state.status.processTime * 1000).toFixed(2)}ms (
        {((state.status.processTime / processTimeLimit) * 100).toFixed(1)}%)
      </div>
    </React.Fragment>
  );
};

window.onload = () => {
  ReactDOM.render(<App />, document.getElementById("root"));
};
window.oncontextmenu = (e: MouseEvent) => {
  e.preventDefault();
  ipcRenderer.send("contextmenu", { x: e.x, y: e.y });
};
