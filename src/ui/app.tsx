import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useState, useEffect, useRef } from "react";
import { Notes } from "./note";

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

const FilterSelect: React.FC = () => {
  const onChange = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "filter", "kind", value]);
  };
  return (
    <select onChange={onChange}>
      <option>none</option>
      <option>lowpass</option>
      <option>highpass</option>
    </select>
  );
};

const FilterFreq: React.FC = () => {
  const onInput = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "filter", "freq", value]);
  };
  return (
    <label>
      Freq
      <input
        onInput={onInput}
        type="range"
        min="30"
        max="20000"
        defaultValue="1000"
      />
    </label>
  );
};

const FilterQ: React.FC = () => {
  const onInput = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "filter", "q", value]);
  };
  return (
    <label>
      Q
      <input
        onInput={onInput}
        type="range"
        min="0.1"
        max="20"
        defaultValue="1"
      />
    </label>
  );
};

const FilterGain: React.FC = () => {
  const onInput = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "filter", "gain", value]);
  };
  return (
    <label>
      Gain
      <input
        onInput={onInput}
        type="range"
        min="-20"
        max="20"
        defaultValue="0"
      />
    </label>
  );
};

const Spectrum = () => {
  const canvasEl: React.MutableRefObject<HTMLCanvasElement | null> = useRef(
    null
  );
  useEffect(() => {
    const callback = (_: any, command: string[]) => {
      if (command[0] === "fft") {
        if (canvasEl.current == null) {
          return;
        }
        const canvas = canvasEl.current;
        const width = canvas.width;
        const height = canvas.height;
        const samples = command.length - 1;
        const sampleRate = 48000;
        const maxFreq = 24000;
        const minFreq = 32;
        const maxDb = -6;
        const minDb = -100;

        const ctx = canvas.getContext("2d")!;
        ctx.fillStyle = "black";
        ctx.fillRect(0, 0, width, height);

        ctx.strokeStyle = "green";
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(0, height);
        for (let i = 0; i < samples; i++) {
          const value = parseFloat(command[i + 1]);
          const freq = (sampleRate / 2) * (i / samples);
          const x =
            (Math.log(freq / minFreq) / Math.log(maxFreq / minFreq)) * width;
          const db = 20 * Math.log10(value);
          const y = (1 - (db - minDb) / (maxDb - minDb)) * height;
          ctx.lineTo(x, y);
        }
        ctx.stroke();
      }
    };
    ipcRenderer.on("audio", callback);
    return () => {
      ipcRenderer.off("audio", callback);
    };
  }, []);
  return <canvas width="512" height="200" ref={canvasEl}></canvas>;
};

const FilterShape = () => {
  const canvasEl: React.MutableRefObject<HTMLCanvasElement | null> = useRef(
    null
  );
  useEffect(() => {
    const callback = (_: any, command: string[]) => {
      if (command[0] === "filter-shape") {
        if (canvasEl.current == null) {
          return;
        }
        const canvas = canvasEl.current;
        const width = canvas.width;
        const height = canvas.height;
        const samples = command.length - 1;
        const sampleRate = 48000;
        const maxFreq = 24000;
        const minFreq = 32;
        const maxDb = 50;
        const minDb = -50;

        const ctx = canvas.getContext("2d")!;
        ctx.fillStyle = "black";
        ctx.fillRect(0, 0, width, height);

        ctx.strokeStyle = "pink";
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(0, height / 2);
        for (let i = 0; i < samples; i++) {
          const value = parseFloat(command[i + 1]);
          const freq = (sampleRate / 2) * (i / samples);
          const x =
            (Math.log(freq / minFreq) / Math.log(maxFreq / minFreq)) * width;
          if (i <= 1) {
            console.log(i, freq, x);
          }
          const db = 20 * Math.log10(value);
          const y = (1 - (db - minDb) / (maxDb - minDb)) * height;
          ctx.lineTo(x, y);
        }
        ctx.stroke();
      }
    };
    ipcRenderer.on("audio", callback);
    return () => {
      ipcRenderer.off("audio", callback);
    };
  }, []);
  return <canvas width="512" height="200" ref={canvasEl}></canvas>;
};

const App = () => {
  const [result, setResult] = useState("");
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
      <h1>Desktop Audio</h1>
      <WaveSelect></WaveSelect>
      <FilterSelect></FilterSelect>
      <FilterFreq></FilterFreq>
      <FilterQ></FilterQ>
      <FilterGain></FilterGain>
      <Notes></Notes>
      <pre>{result}</pre>
      <Spectrum />
      <FilterShape />
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
