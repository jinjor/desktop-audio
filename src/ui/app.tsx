import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useState, useEffect, useRef } from "react";
import { Notes } from "./note";
import { Knob } from "./knob";

const MonoPolySelect: React.FC = () => {
  const onChange = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", [value]);
  };
  return (
    <select onChange={onChange}>
      <option>mono</option>
      <option>poly</option>
    </select>
  );
};

const WaveSelect: React.FC = () => {
  const onChange = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "osc", "kind", value]);
  };
  return (
    <select onChange={onChange}>
      <option>sine</option>
      <option>triangle</option>
      <option>square</option>
      <option>pluse</option>
      <option>saw</option>
      <option>noise</option>
    </select>
  );
};

const Attack: React.FC = () => {
  const [value, setValue] = useState(10);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "attack", value]);
    setValue(value);
  };
  return (
    <div>
      <Knob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={value}
        onInput={onInput}
      />
      <label>Attack</label>
    </div>
  );
};
const Decay: React.FC = () => {
  const [value, setValue] = useState(100);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "decay", value]);
    setValue(value);
  };
  return (
    <div>
      <Knob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={value}
        onInput={onInput}
      />
      <label>Decay</label>
    </div>
  );
};

const Sustain: React.FC = () => {
  const [value, setValue] = useState(0.7);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "sustain", value]);
    setValue(value);
  };
  return (
    <div>
      <Knob
        min={0}
        max={1}
        steps={400}
        exponential={false}
        value={value}
        onInput={onInput}
      />
      <label>Sustain</label>
    </div>
  );
};

const Release: React.FC = () => {
  const [value, setValue] = useState(200);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "adsr", "release", value]);
    setValue(value);
  };
  return (
    <div>
      <Knob
        min={0}
        max={800}
        steps={400}
        exponential={true}
        value={value}
        onInput={onInput}
      />
      <label>Release</label>
    </div>
  );
};

const FilterSelect: React.FC = () => {
  const onChange = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "filter", "kind", value]);
  };
  return (
    <div>
      <select onChange={onChange}>
        <option>none</option>
        <option>lowpass-fir</option>
        <option>highpass-fir</option>
        <option>lowpass</option>
        <option>highpass</option>
        <option>bandpass-1</option>
        <option>bandpass-2</option>
        <option>notch</option>
        <option>peaking</option>
        <option>lowshelf</option>
        <option>highshelf</option>
      </select>
    </div>
  );
};

const FilterFreq: React.FC = () => {
  const [value, setValue] = useState(1000);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "freq", value]);
    setValue(value);
  };
  return (
    <div>
      <Knob
        min={30}
        max={20000}
        steps={400}
        exponential={true}
        value={value}
        onInput={onInput}
      />
      <label>Freq</label>
    </div>
  );
};

const FilterQ: React.FC = () => {
  const [value, setValue] = useState(0);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "q", value]);
    setValue(value);
  };
  return (
    <div>
      <Knob
        min={0}
        max={20}
        steps={400}
        exponential={false}
        value={value}
        onInput={onInput}
      />
      <label>Q</label>
    </div>
  );
};

const FilterGain: React.FC = () => {
  const [value, setValue] = useState(0);
  const onInput = (value: number) => {
    ipcRenderer.send("audio", ["set", "filter", "gain", value]);
    setValue(value);
  };
  return (
    <div>
      <Knob
        min={-40}
        max={40}
        steps={400}
        exponential={false}
        value={value}
        onInput={onInput}
      />
      <label>Gain</label>
    </div>
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
      <MonoPolySelect />
      <WaveSelect />
      <div style={{ display: "flex" }}>
        <div>
          <Attack />
          <Decay />
          <Sustain />
          <Release />
        </div>
        <div>
          <FilterSelect />
          <FilterFreq />
          <FilterQ />
          <FilterGain />
        </div>
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
