import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useState, useEffect, useRef, useCallback } from "react";
import { Notes } from "./note";
import { LabeledKnob } from "./knob";
import { Radio } from "./radio";
import * as d from "./decoder";

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
  (o: { onChange: (value: string) => void; value: string }) => {
    return (
      <Radio list={["mono", "poly"]} value={o.value} onChange={o.onChange} />
    );
  }
);
const GlideTime = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
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
        onChange={o.onChange}
        label="Glide"
      />
    );
  }
);
const OscKind = React.memo(
  (o: { onChange: (value: string) => void; value: string }) => {
    return (
      <Radio
        list={[
          "sine",
          "triangle",
          "square",
          "square-wt",
          "pulse",
          "saw",
          "saw-wt",
          "saw-rev",
        ]}
        value={o.value}
        onChange={o.onChange}
      />
    );
  }
);
const Octave = React.memo(
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
const Coarse = React.memo(
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
const Fine = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
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
        onChange={o.onChange}
        label="Fine"
      />
    );
  }
);
const Attack = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    return (
      <LabeledKnob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={o.onChange}
        label="Attack"
      />
    );
  }
);
const Decay = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    return (
      <LabeledKnob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={o.onChange}
        label="Decay"
      />
    );
  }
);
const Sustain = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    return (
      <LabeledKnob
        min={0}
        max={1}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Sustain"
      />
    );
  }
);
const Release = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    return (
      <LabeledKnob
        min={0}
        max={800}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={o.onChange}
        label="Release"
      />
    );
  }
);
const FilterKind = React.memo(
  (o: { onChange: (value: string) => void; value: string }) => {
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
        onChange={o.onChange}
      />
    );
  }
);
const FilterFreq = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    return (
      <LabeledKnob
        min={30}
        max={20000}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={o.onChange}
        label="Freq"
      />
    );
  }
);
const FilterQ = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    return (
      <LabeledKnob
        min={0}
        max={20}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Q"
      />
    );
  }
);
const FilterGain = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    return (
      <LabeledKnob
        min={-40}
        max={40}
        steps={400}
        exponential={false}
        value={o.value}
        from={0}
        onChange={o.onChange}
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
  (o: {
    index: number;
    value: LFO;
    onChangeDestination: (index: number, value: string) => void;
    onChangeWave: (index: number, value: string) => void;
    onChangeFreq: (index: number, value: number) => void;
    onChangeAmount: (index: number, value: number) => void;
  }) => {
    const list = [...lfoDestinations.keys()];
    const destination = lfoDestinations.get(o.value.destination)!;
    const onChangeDestination = useCallbackWithIndex(
      o.index,
      o.onChangeDestination
    );
    const onChangeWave = useCallbackWithIndex(o.index, o.onChangeWave);
    const onChangeFreq = useCallbackWithIndex(o.index, o.onChangeFreq);
    const onChangeAmount = useCallbackWithIndex(o.index, o.onChangeAmount);
    return (
      <EditGroup label={`LFO ${o.index + 1}`}>
        <div style={{ display: "flex", gap: "12px" }}>
          <div>
            <Radio
              list={list}
              value={o.value.destination}
              onChange={onChangeDestination}
            />
          </div>
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
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
          "square",
          "square-wt",
          "pulse",
          "saw",
          "saw-wt",
          "saw-rev",
        ]}
        value={o.value}
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
  filter: d.object({
    kind: d.string(),
    freq: d.number(),
    q: d.number(),
    gain: d.number(),
  }),
});
const App = () => {
  useEffect(() => {
    const callback = (_: any, command: string[]) => {
      if (command[0] === "all_params") {
        const obj = JSON.parse(command[1]);
        const state = stateDecoder.run(obj.state);
        setPoly(state.poly);
        setGlideTime(state.glideTime);
        setOscKind(state.osc.kind);
        setOctave(state.osc.octave);
        setCoarse(state.osc.coarse);
        setFine(state.osc.fine);
        setAttack(state.adsr.attack);
        setDecay(state.adsr.decay);
        setSustain(state.adsr.sustain);
        setRelease(state.adsr.release);
        setLFOs(state.lfos);
        setFilterKind(state.filter.kind);
        setFilterFreq(state.filter.freq);
        setFilterQ(state.filter.q);
        setFilterGain(state.filter.gain);
      }
    };
    ipcRenderer.on("audio", callback);
    return () => {
      ipcRenderer.off("audio", callback);
    };
  }, []);
  const [poly, setPoly] = useState("mono");
  const [glideTime, setGlideTime] = useState(100);
  const onChangePoly = useCallback((value: string) => {
    ipcRenderer.send("audio", [value]);
    setPoly(value);
  }, []);
  const onChangeGlideTime = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "glide_time", value]);
    setGlideTime(value);
  }, []);

  const [oscKind, setOscKind] = useState("sine");
  const [octave, setOctave] = useState(0);
  const [coarse, setCoarse] = useState(0);
  const [fine, setFine] = useState(0);
  const onChangeOscKind = useCallback((value: string) => {
    ipcRenderer.send("audio", ["set", "osc", "kind", value]);
    setOscKind(value);
  }, []);
  const onChangeOctave = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "osc", "octave", value]);
    setOctave(value);
  }, []);
  const onChangeCoarse = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "osc", "coarse", value]);
    setCoarse(value);
  }, []);
  const onChangeFine = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "osc", "fine", value]);
    setFine(value);
  }, []);

  const [attack, setAttack] = useState(10);
  const [decay, setDecay] = useState(100);
  const [sustain, setSustain] = useState(0.7);
  const [release, setRelease] = useState(200);
  const onChangeAttack = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "attack", value]);
    setAttack(value);
  }, []);
  const onChangeDecay = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "decay", value]);
    setDecay(value);
  }, []);
  const onChangeSustain = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "sustain", value]);
    setSustain(value);
  }, []);
  const onChangeRelease = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "release", value]);
    setRelease(value);
  }, []);

  const [lfos, setLFOs] = useState([defaultLFO, defaultLFO, defaultLFO]);
  const onChangeLFODestination = useCallback(
    (i: number, value: string) => {
      const { defaultFreq, defaultAmount } = lfoDestinations.get(value)!;
      setLFOs(
        setItem(lfos, i, {
          destination: value,
          freq: defaultFreq,
          amount: defaultAmount,
        })
      );
      ipcRenderer.send("audio", [
        "set",
        "lfo",
        String(i),
        "destination",
        value,
      ]);
      ipcRenderer.send("audio", ["set", "lfo", String(i), "freq", defaultFreq]);
      ipcRenderer.send("audio", [
        "set",
        "lfo",
        String(i),
        "amount",
        defaultAmount,
      ]);
    },
    [lfos]
  );
  const onChangeLFOWave = useCallback(
    (i: number, wave: string) => {
      ipcRenderer.send("audio", ["set", "lfo", String(i), "wave", wave]);
      setLFOs(setItem(lfos, i, { wave }));
    },
    [lfos]
  );
  const onChangeLFOFreq = useCallback(
    (i: number, freq: number) => {
      ipcRenderer.send("audio", ["set", "lfo", String(i), "freq", freq]);
      setLFOs(setItem(lfos, i, { freq }));
    },
    [lfos]
  );
  const onChangeLFOAmount = useCallback(
    (i: number, amount: number) => {
      ipcRenderer.send("audio", ["set", "lfo", String(i), "amount", amount]);
      setLFOs(setItem(lfos, i, { amount }));
    },
    [lfos]
  );

  const [filterKind, setFilterKind] = useState("none");
  const [filterFreq, setFilterFreq] = useState(1000);
  const [filterQ, setFilterQ] = useState(0);
  const [filterGain, setFilterGain] = useState(0);
  const onChangeFilterKind = useCallback((value: string) => {
    ipcRenderer.send("audio", ["set", "filter", "kind", value]);
    setFilterKind(value);
  }, []);
  const onChangeFilterFreq = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "freq", value]);
    setFilterFreq(value);
  }, []);
  const onChangeFilterQ = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "q", value]);
    setFilterQ(value);
  }, []);
  const onChangeFilterGain = useCallback((value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "gain", value]);
    setFilterGain(value);
  }, []);
  return (
    <React.Fragment>
      <div style={{ display: "flex", gap: "20px", padding: "5px 10px" }}>
        <EditGroup label="POLY">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <Poly value={poly} onChange={onChangePoly} />
            <GlideTime value={glideTime} onChange={onChangeGlideTime} />
          </div>
        </EditGroup>
        <EditGroup label="OSC">
          <div style={{ display: "flex", gap: "12px" }}>
            <div>
              <OscKind value={oscKind} onChange={onChangeOscKind} />
            </div>
            <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
              <Octave value={octave} onChange={onChangeOctave} />
              <Coarse value={coarse} onChange={onChangeCoarse} />
              <Fine value={fine} onChange={onChangeFine} />
            </div>
          </div>
        </EditGroup>
        <EditGroup label="Envelope">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <Attack value={attack} onChange={onChangeAttack} />
            <Decay value={decay} onChange={onChangeDecay} />
            <Sustain value={sustain} onChange={onChangeSustain} />
            <Release value={release} onChange={onChangeRelease} />
          </div>
        </EditGroup>
        <EditGroup label="FILTER">
          <div style={{ display: "flex", gap: "12px" }}>
            <div>
              <FilterKind onChange={onChangeFilterKind} value={filterKind} />
            </div>
            <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
              <FilterFreq onChange={onChangeFilterFreq} value={filterFreq} />
              <FilterQ onChange={onChangeFilterQ} value={filterQ} />
              <FilterGain onChange={onChangeFilterGain} value={filterGain} />
            </div>
          </div>
        </EditGroup>
        <LFOGroup
          index={0}
          value={lfos[0]}
          onChangeDestination={onChangeLFODestination}
          onChangeWave={onChangeLFOWave}
          onChangeFreq={onChangeLFOFreq}
          onChangeAmount={onChangeLFOAmount}
        />
        <LFOGroup
          index={1}
          value={lfos[1]}
          onChangeDestination={onChangeLFODestination}
          onChangeWave={onChangeLFOWave}
          onChangeFreq={onChangeLFOFreq}
          onChangeAmount={onChangeLFOAmount}
        />
        <LFOGroup
          index={2}
          value={lfos[2]}
          onChangeDestination={onChangeLFODestination}
          onChangeWave={onChangeLFOWave}
          onChangeFreq={onChangeLFOFreq}
          onChangeAmount={onChangeLFOAmount}
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
