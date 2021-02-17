import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useEffect, useRef, useCallback, useReducer } from "react";
import { Notes } from "./note";
import { LabeledKnob } from "./knob";
import { Radio } from "./radio";
import { Select } from "./select";
import * as d from "./decoder";
import * as waveform from "./waveform";

const waveNameToIcon = (value: string) => {
  const size = 16;
  switch (value) {
    case "sine":
      return waveform.sine(size);
    case "triangle":
      return waveform.triangle(size);
    case "square":
    case "square-wt":
      return waveform.square(size);
    case "pulse":
      return waveform.pulse(size);
    case "saw":
    case "saw-wt":
      return waveform.saw(size);
    case "saw-rev":
      return waveform.sawRev(size);
    case "noise":
      return waveform.noise(size);
    default:
      return value;
  }
};

const EditGroup = (o: { label: string; children: any }) => {
  return (
    <div style={{ display: "flex", flexFlow: "column" }}>
      <label
        style={{
          display: "block",
          borderBottom: "solid 1px #aaa",
          textAlign: "center",
        }}
      >
        {o.label}
      </label>
      <div style={{ padding: "5px 0" }}>{o.children}</div>
    </div>
  );
};

const Poly = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedPoly", value });
    return (
      <Radio list={["mono", "poly"]} value={o.value} onChange={onChange} />
    );
  }
);
const GlideTime = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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

const OscKind = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedOscKind", value });
    return (
      <Radio
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
        columns={2}
        toElement={waveNameToIcon}
        onChange={onChange}
      />
    );
  }
);
const Octave = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedOscOctave", value });
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
const Coarse = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedOscCoarse", value });
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
const Fine = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedOscFine", value });
    const min = -100;
    const max = 100;
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
        label="Fine"
      />
    );
  }
);
const Attack = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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
const FilterKind = React.memo(
  (o: { dispatch: React.Dispatch<Action>; value: string }) => {
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
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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
  (o: { dispatch: React.Dispatch<Action>; value: number }) => {
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
type LFO = {
  destination: string;
  wave: string;
  freq: number;
  amount: number;
};
const defaultLFO = {
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
  (o: { index: number; value: LFO; dispatch: React.Dispatch<Action> }) => {
    const list = [...lfoDestinations.keys()];
    const destination = lfoDestinations.get(o.value.destination)!;
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
      <EditGroup label={`LFO ${o.index + 1}`}>
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <Select
            list={list}
            value={o.value.destination}
            onChange={onChangeDestination}
          />
          <LFOWave value={o.value.wave} onChange={onChangeWave} />
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
      <Radio
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
        columns={2}
        toElement={waveNameToIcon}
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
  destination: string;
  delay: number;
  attack: number;
  amount: number;
};
const defaultEnvelope: Envelope = {
  destination: "none",
  delay: 0,
  attack: 0,
  amount: 0,
};
const EnvelopeGroup = React.memo(
  (o: { index: number; value: Envelope; dispatch: React.Dispatch<Action> }) => {
    const list = [...envelopeDestinations.keys()];
    const destination = envelopeDestinations.get(o.value.destination)!;
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
      <EditGroup label={`Envelope ${o.index + 1}`}>
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

const stateDecoder = d.object({
  poly: d.string(),
  glideTime: d.number(),
  osc: d.object({
    kind: d.string(),
    octave: d.number(),
    coarse: d.number(),
    fine: d.number(),
  }),
  adsr: d.object({
    attack: d.number(),
    decay: d.number(),
    sustain: d.number(),
    release: d.number(),
  }),
  lfos: d.array(
    d.object({
      destination: d.string(),
      wave: d.string(),
      freq: d.number(),
      amount: d.number(),
    })
  ),
  envelopes: d.array(
    d.object({
      destination: d.string(),
      delay: d.number(),
      attack: d.number(),
      amount: d.number(),
    })
  ),
  filter: d.object({
    kind: d.string(),
    freq: d.number(),
    q: d.number(),
    gain: d.number(),
  }),
});
type State = d.TypeOf<typeof stateDecoder>;
type Action =
  | { type: "receivedCommand"; command: string[] }
  | { type: "changedPoly"; value: string }
  | { type: "changedGlideTime"; value: number }
  | { type: "changedOscKind"; value: string }
  | { type: "changedOscOctave"; value: number }
  | { type: "changedOscCoarse"; value: number }
  | { type: "changedOscFine"; value: number }
  | { type: "changedAdsrAttack"; value: number }
  | { type: "changedAdsrDecay"; value: number }
  | { type: "changedAdsrSustain"; value: number }
  | { type: "changedAdsrRelease"; value: number }
  | { type: "changedLFODestination"; index: number; value: string }
  | { type: "changedLFOWave"; index: number; value: string }
  | { type: "changedLFOFreq"; index: number; value: number }
  | { type: "changedLFOAmount"; index: number; value: number }
  | { type: "changedEnvelopeDestination"; index: number; value: string }
  | { type: "changedEnvelopeDelay"; index: number; value: number }
  | { type: "changedEnvelopeAttack"; index: number; value: number }
  | { type: "changedEnvelopeAmount"; index: number; value: number }
  | { type: "changedFilterKind"; value: string }
  | { type: "changedFilterFreq"; value: number }
  | { type: "changedFilterQ"; value: number }
  | { type: "changedFilterGain"; value: number };

const App = () => {
  const initialState: State = {
    poly: "mono",
    glideTime: 100,
    osc: {
      kind: "sine",
      octave: 0,
      coarse: 0,
      fine: 0,
    },
    adsr: {
      attack: 0,
      decay: 100,
      sustain: 0.7,
      release: 100,
    },
    lfos: [defaultLFO, defaultLFO, defaultLFO],
    envelopes: [defaultEnvelope, defaultEnvelope, defaultEnvelope],
    filter: {
      kind: "none",
      freq: 1000,
      q: 0,
      gain: 0,
    },
  };
  const [state, dispatch] = useReducer((state: State, action: Action) => {
    switch (action.type) {
      case "receivedCommand": {
        const { command } = action;
        if (command[0] === "all_params") {
          const obj = JSON.parse(command[1]);
          console.log(obj);
          return stateDecoder.run(obj.state);
        } else {
          return state;
        }
      }
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
      case "changedOscKind": {
        const { value } = action;
        ipcRenderer.send("audio", ["set", "osc", "kind", value]);
        return { ...state, osc: { ...state.osc, kind: value } };
      }
      case "changedOscOctave": {
        const { value } = action;
        ipcRenderer.send("audio", ["set", "osc", "octave", value]);
        return { ...state, osc: { ...state.osc, octave: value } };
      }
      case "changedOscCoarse": {
        const { value } = action;
        ipcRenderer.send("audio", ["set", "osc", "coarse", value]);
        return { ...state, osc: { ...state.osc, coarse: value } };
      }
      case "changedOscFine": {
        const { value } = action;
        ipcRenderer.send("audio", ["set", "osc", "fine", value]);
        return { ...state, osc: { ...state.osc, fine: value } };
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
    }
  }, initialState);
  useEffect(() => {
    const callback = (_: any, command: string[]) => {
      dispatch({ type: "receivedCommand", command });
    };
    ipcRenderer.on("audio", callback);
    return () => {
      ipcRenderer.off("audio", callback);
    };
  }, []);
  return (
    <React.Fragment>
      <div style={{ display: "flex", gap: "20px", padding: "5px 10px" }}>
        <EditGroup label="POLY">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <Poly value={state.poly} dispatch={dispatch} />
            <GlideTime value={state.glideTime} dispatch={dispatch} />
          </div>
        </EditGroup>
        <EditGroup label="OSC">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <OscKind value={state.osc.kind} dispatch={dispatch} />
            <Octave value={state.osc.octave} dispatch={dispatch} />
            <Coarse value={state.osc.coarse} dispatch={dispatch} />
            <Fine value={state.osc.fine} dispatch={dispatch} />
          </div>
        </EditGroup>
        <EditGroup label="Envelope">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <Attack value={state.adsr.attack} dispatch={dispatch} />
            <Decay value={state.adsr.decay} dispatch={dispatch} />
            <Sustain value={state.adsr.sustain} dispatch={dispatch} />
            <Release value={state.adsr.release} dispatch={dispatch} />
          </div>
        </EditGroup>
        <EditGroup label="FILTER">
          <div style={{ display: "flex", gap: "12px" }}>
            <div>
              <FilterKind dispatch={dispatch} value={state.filter.kind} />
            </div>
            <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
              <FilterFreq dispatch={dispatch} value={state.filter.freq} />
              <FilterQ dispatch={dispatch} value={state.filter.q} />
              <FilterGain dispatch={dispatch} value={state.filter.gain} />
            </div>
          </div>
        </EditGroup>
        <LFOGroup index={0} value={state.lfos[0]} dispatch={dispatch} />
        <LFOGroup index={1} value={state.lfos[1]} dispatch={dispatch} />
        <LFOGroup index={2} value={state.lfos[2]} dispatch={dispatch} />
        <EnvelopeGroup
          index={0}
          value={state.envelopes[0]}
          dispatch={dispatch}
        />
        <EnvelopeGroup
          index={1}
          value={state.envelopes[1]}
          dispatch={dispatch}
        />
        <EnvelopeGroup
          index={2}
          value={state.envelopes[2]}
          dispatch={dispatch}
        />
      </div>
      <Notes />
      <Spectrum />
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
