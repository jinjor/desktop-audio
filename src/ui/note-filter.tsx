import React from "react";
import { ParamGroup } from "./param-group";
import { LabeledKnob } from "./knob";
import { Select } from "./select";
import { ParamsAction } from "./state";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

export const noteFilterDecoder = d.object({
  enabled: d.boolean,
  targetOsc: d.string(),
  kind: d.string(),
  octave: d.number(),
  coarse: d.number(),
  q: d.number(),
  gain: d.number(),
});
type NoteFilter = d.TypeOf<typeof noteFilterDecoder>;

export const NoteFilterGroup = React.memo(
  (o: { value: NoteFilter; dispatch: React.Dispatch<ParamsAction> }) => {
    checkRenderingExclusive("params", "note_filter");
    const onChangeNoteFilterEnabled = (value: boolean) => {
      o.dispatch({
        type: "changedNoteFilterEnabled",
        value,
      });
    };
    return (
      <ParamGroup
        label="Note Filter"
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeNoteFilterEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <NoteFilterTargetOsc
            dispatch={o.dispatch}
            value={o.value.targetOsc}
          />
          <NoteFilterKind dispatch={o.dispatch} value={o.value.kind} />
          <NoteFilterOctave dispatch={o.dispatch} value={o.value.octave} />
          <NoteFilterCoarse dispatch={o.dispatch} value={o.value.coarse} />
          <NoteFilterQ dispatch={o.dispatch} value={o.value.q} />
          <NoteFilterGain dispatch={o.dispatch} value={o.value.gain} />
        </div>
      </ParamGroup>
    );
  }
);

const NoteFilterKind = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedNoteFilterKind", value });
    return (
      <Select
        list={[
          "none",
          "lowpass",
          "highpass",
          "bandpass-1",
          "bandpass-2",
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
const NoteFilterTargetOsc = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const theirList = ["all", "0", "1"];
    const ourList = ["OSC 1 & 2", "OSC 1", "OSC 2"];
    const onChange = (value: string) =>
      o.dispatch({
        type: "changedNoteFilterTargetOsc",
        value: theirList[ourList.indexOf(value)],
      });
    return (
      <Select
        list={ourList}
        value={ourList[theirList.indexOf(o.value)]}
        onChange={onChange}
      />
    );
  }
);
const NoteFilterOctave = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({
        type: "changedNoteFilterOctave",
        value,
      });
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
const NoteFilterCoarse = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({
        type: "changedNoteFilterCoarse",
        value,
      });
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
const NoteFilterQ = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedNoteFilterQ", value });
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
const NoteFilterGain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedNoteFilterGain", value });
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
