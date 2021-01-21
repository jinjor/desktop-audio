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

const Canvas = (props: {
  listen: (canvas: HTMLCanvasElement) => () => void;
  [key: string]: any;
}) => {
  const { listen, ...canvasProps } = props;
  const canvasEl: React.MutableRefObject<HTMLCanvasElement | null> = useRef(
    null
  );
  useEffect(() => {
    return listen(canvasEl.current!);
  }, []);
  return <canvas {...canvasProps} ref={canvasEl}></canvas>;
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
