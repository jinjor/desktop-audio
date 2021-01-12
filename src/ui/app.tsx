import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useState, useEffect, useRef } from "react";
import { Notes } from "./note";

const WaveSelect: React.FC = () => {
  const onChange = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "kind", value]);
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

const App = () => {
  const [result, setResult] = useState("");
  const canvasEl: React.MutableRefObject<HTMLCanvasElement | null> = useRef(
    null
  );
  useEffect(() => {
    ipcRenderer.on("audio", (_: any, command: string[]) => {
      setResult(JSON.stringify(command));
      if (command[0] === "fft") {
        const canvas = canvasEl.current!;
        const ctx = canvas.getContext("2d")!;
        ctx.fillStyle = "black";
        ctx.fillRect(0, 0, 128, 100);
        ctx.strokeStyle = "green";
        ctx.beginPath();
        ctx.moveTo(0, 100);
        for (let i = 1; i < command.length; i++) {
          const value = parseFloat(command[i]);
          ctx.lineTo(i, (1 - value) * 100);
        }
        ctx.closePath();
        ctx.stroke();
      }
    });
  }, []);
  return (
    <React.Fragment>
      <h1>Desktop Audio</h1>
      <WaveSelect></WaveSelect>
      <Notes></Notes>
      <pre>{result}</pre>
      <canvas width="128" height="100" ref={canvasEl}></canvas>
    </React.Fragment>
  );
};

window.onload = () => {
  ReactDOM.render(<App />, document.getElementById("root"));
};
