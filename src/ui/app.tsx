import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useState, useEffect, useRef } from "react";
import { Notes } from "./note";
import { LabeledKnob } from "./knob";
import { Radio } from "./radio";

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

const MonoPolySelect = () => {
  const [value, setValue] = useState("mono");
  const onChange = (value: string) => {
    setValue(value);
    ipcRenderer.send("audio", [value]);
  };
  return <Radio list={["mono", "poly"]} value={value} onChange={onChange} />;
};
const GlideTime = () => {
  const [value, setValue] = useState(100);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "glide_time", value]);
    setValue(value);
  };
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
      value={value}
      onChange={onInput}
      label="Glide"
    />
  );
};
const WaveSelect = () => {
  const [value, setValue] = useState("sine");
  const onChange = (value: string) => {
    setValue(value);
    ipcRenderer.send("audio", ["set", "osc", "kind", value]);
  };
  return (
    <Radio
      list={["sine", "triangle", "square", "pulse", "saw", "noise"]}
      value={value}
      onChange={onChange}
    />
  );
};
const Octave = () => {
  const [value, setValue] = useState(0);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "osc", "octave", value]);
    setValue(value);
  };
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
      value={value}
      onChange={onInput}
      label="Octave"
    />
  );
};
const Coarse = () => {
  const [value, setValue] = useState(0);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "osc", "coarse", value]);
    setValue(value);
  };
  const min = -12;
  const max = 12;
  const steps = max - min + 1;
  return (
    <LabeledKnob
      min={min}
      max={max}
      steps={steps}
      exponential={true}
      value={value}
      from={0}
      onChange={onInput}
      label="Coarse"
    />
  );
};
const Fine = () => {
  const [value, setValue] = useState(0);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "osc", "fine", value]);
    setValue(value);
  };
  const min = -100;
  const max = 100;
  const steps = max - min + 1;
  return (
    <LabeledKnob
      min={min}
      max={max}
      steps={steps}
      exponential={true}
      value={value}
      from={0}
      onChange={onInput}
      label="Fine"
    />
  );
};
const Attack = () => {
  const [value, setValue] = useState(10);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "attack", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={0}
      max={400}
      steps={400}
      exponential={true}
      value={value}
      onChange={onInput}
      label="Attack"
    />
  );
};
const Decay = () => {
  const [value, setValue] = useState(100);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "decay", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={0}
      max={400}
      steps={400}
      exponential={true}
      value={value}
      onChange={onInput}
      label="Decay"
    />
  );
};

const Sustain = () => {
  const [value, setValue] = useState(0.7);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "sustain", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={0}
      max={1}
      steps={400}
      exponential={false}
      value={value}
      onChange={onInput}
      label="Sustain"
    />
  );
};

const Release = () => {
  const [value, setValue] = useState(200);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "release", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={0}
      max={800}
      steps={400}
      exponential={true}
      value={value}
      onChange={onInput}
      label="Release"
    />
  );
};

const FilterSelect = () => {
  const [value, setValue] = useState("none");
  const onChange = (value: string) => {
    setValue(value);
    ipcRenderer.send("audio", ["set", "filter", "kind", value]);
  };
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
      value={value}
      onChange={onChange}
    />
  );
};

const FilterFreq = () => {
  const [value, setValue] = useState(1000);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "freq", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={30}
      max={20000}
      steps={400}
      exponential={true}
      value={value}
      onChange={onInput}
      label="Freq"
    />
  );
};

const FilterQ = () => {
  const [value, setValue] = useState(0);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "q", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={0}
      max={20}
      steps={400}
      exponential={false}
      value={value}
      onChange={onInput}
      label="Q"
    />
  );
};

const FilterGain = () => {
  const [value, setValue] = useState(0);
  const onChange = (value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "gain", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={-40}
      max={40}
      steps={400}
      exponential={false}
      value={value}
      from={0}
      onChange={onChange}
      label="Gain"
    />
  );
};

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
lfoDestinations.set("vibrato-exp", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 40,
  defaultFreq: 10,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("fm", {
  freqType: "ratio",
  minFreq: 0.1,
  maxFreq: 10,
  defaultFreq: 3,
  minAmount: 0,
  maxAmount: 1.57,
  defaultAmount: 1,
  fromAmount: 0,
});
const LFO = (o: { index: number }) => {
  const list = [...lfoDestinations.keys()];
  const [value, setValue] = useState("none");
  const [destination, setDestination] = useState(lfoDestinations.get("none")!);
  const onChange = (value: string) => {
    const destination = lfoDestinations.get(value)!;
    ipcRenderer.send("audio", [
      "set",
      "lfo",
      String(o.index),
      "destination",
      value,
    ]);
    setValue(value);
    setDestination(destination);
  };
  return (
    <EditGroup label={`LFO ${o.index + 1}`}>
      <div style={{ display: "flex", gap: "12px" }}>
        <div>
          <Radio list={list} value={value} onChange={onChange} />
        </div>
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <LFOWaveSelect index={o.index} />
          <LFOFreq
            index={o.index}
            min={destination.minFreq}
            max={destination.maxFreq}
            value={destination.defaultFreq}
          />
          <LFOAmount
            index={o.index}
            min={destination.minAmount}
            max={destination.maxAmount}
            value={destination.defaultAmount}
            from={destination.fromAmount}
          />
        </div>
      </div>
    </EditGroup>
  );
};
const LFOWaveSelect = (o: { index: number }) => {
  const [value, setValue] = useState("sine");
  const onChange = (value: string) => {
    setValue(value);
    ipcRenderer.send("audio", ["set", "lfo", String(o.index), "wave", value]);
  };
  return (
    <Radio
      list={["sine", "triangle", "square", "pulse", "saw", "saw-rev"]}
      value={value}
      onChange={onChange}
    />
  );
};
const LFOFreq = (o: {
  index: number;
  min: number;
  max: number;
  value: number;
}) => {
  const [value, setValue] = useState(o.value);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "lfo", String(o.index), "freq", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={o.min}
      max={o.max}
      steps={400}
      exponential={false}
      value={value}
      onChange={onInput}
      label="Freq"
    />
  );
};
const LFOAmount = (o: {
  index: number;
  min: number;
  max: number;
  value: number;
  from: number;
}) => {
  const [value, setValue] = useState(o.value);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "lfo", String(o.index), "amount", value]);
    setValue(value);
  };
  return (
    <LabeledKnob
      min={o.min}
      max={o.max}
      steps={400}
      exponential={false}
      value={value}
      from={o.from}
      onChange={onInput}
      label="Amount"
    />
  );
};
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

const App = () => {
  const [, setResult] = useState("");
  useEffect(() => {
    const callback = (_: any, command: string[]) => {
      setResult(JSON.stringify(command));
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
            <MonoPolySelect />
            <GlideTime />
          </div>
        </EditGroup>
        <EditGroup label="OSC">
          <div style={{ display: "flex", gap: "12px" }}>
            <div>
              <WaveSelect />
            </div>
            <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
              <Octave />
              <Coarse />
              <Fine />
            </div>
          </div>
        </EditGroup>
        <EditGroup label="Envelope">
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <Attack />
            <Decay />
            <Sustain />
            <Release />
          </div>
        </EditGroup>
        <EditGroup label="FILTER">
          <div style={{ display: "flex", gap: "12px" }}>
            <div>
              <FilterSelect />
            </div>
            <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
              <FilterFreq />
              <FilterQ />
              <FilterGain />
            </div>
          </div>
        </EditGroup>
        <LFO index={0} />
        <LFO index={1} />
        <LFO index={2} />
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
