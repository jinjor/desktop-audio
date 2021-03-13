import { ipcRenderer } from "electron";
import ReactDOM from "react-dom";
import React, { useEffect, useRef, useCallback } from "react";
import { Notes } from "./note";
import { initialState, ParamsAction, reducer } from "./state";
import { OSCGroup } from "./osc";
import { LFOGroup } from "./lfo";
import { ADSRGroup } from "./adsr";
import { NoteFilterGroup } from "./note-filter";
import { FilterGroup } from "./filter";
import { EnvelopeGroup } from "./envelope";
import { EchoGroup } from "./echo";
import { ModeGroup } from "./mode";
import { FormantGroup } from "./formant";
import { PresetAction, Presets } from "./preset";
import { useReducerWithEffect } from "./react-util";

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
  const [state, dispatch] = useReducerWithEffect(reducer, initialState);
  useEffect(() => {
    const callback = (_: any, command: string[]) => {
      dispatch({ type: "receivedCommand", command });
    };
    ipcRenderer.on("audio", callback);
    return () => {
      ipcRenderer.off("audio", callback);
    };
  }, []);
  const dispatchParam: React.Dispatch<ParamsAction> = useCallback(
    (action: ParamsAction) => {
      dispatch({
        type: "paramsAction",
        value: action,
      });
    },
    []
  );
  const dispatchPreset: React.Dispatch<PresetAction> = useCallback(
    (action: PresetAction) => {
      dispatch({
        type: "presetAction",
        value: action,
      });
    },
    []
  );
  const p = state.params;
  const processTimeLimit = 0.0213; // TODO: get this from server
  return (
    <React.Fragment>
      <div>
        <Presets state={state.preset} dispatch={dispatchPreset} />
      </div>
      <div style={{ display: "flex", gap: "20px", padding: "5px 10px" }}>
        <ModeGroup
          poly={p.poly}
          glideTime={p.glideTime}
          velSense={p.velSense}
          dispatch={dispatchParam}
        />
        <OSCGroup index={0} value={p.oscs[0]} dispatch={dispatchParam} />
        <OSCGroup index={1} value={p.oscs[1]} dispatch={dispatchParam} />
        <ADSRGroup value={p.adsr} dispatch={dispatchParam} />
        <NoteFilterGroup value={p.noteFilter} dispatch={dispatchParam} />
        <FilterGroup value={p.filter} dispatch={dispatchParam} />
        <FormantGroup value={p.formant} dispatch={dispatchParam} />
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
        <EchoGroup value={p.echo} dispatch={dispatchParam} />
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
