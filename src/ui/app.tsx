import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useState, useEffect } from "react";
import { Notes } from "./note";

const WaveSelect: React.FC = () => {
  const onChange = (e: any) => {
    const value = e.target.value;
    ipcRenderer.send("audio", ["set", "kind", value]);
  };
  return (
    <select onChange={onChange}>
      <option>sine</option>
      <option>square</option>
      <option>saw</option>
      <option>noise</option>
    </select>
  );
};

const App = () => {
  const [result, setResult] = useState("");
  useEffect(() => {
    ipcRenderer.on("audio", (_: any, command: string[]) => {
      setResult(JSON.stringify(command));
    });
  }, []);
  return (
    <React.Fragment>
      <h1>Desktop Audio</h1>
      <WaveSelect></WaveSelect>
      <Notes></Notes>
      <pre>{result}</pre>
    </React.Fragment>
  );
};

window.onload = () => {
  ReactDOM.render(<App />, document.getElementById("root"));
};
