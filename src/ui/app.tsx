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
  const onInput = (value: number) => {
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
      onChange={onInput}
      label="Gain"
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
  return <Canvas width="512" height="200" listen={listen} />;
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
          <MonoPolySelect />
        </EditGroup>
        <EditGroup label="WAVE">
          <WaveSelect />
        </EditGroup>
        <EditGroup label="EG">
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
