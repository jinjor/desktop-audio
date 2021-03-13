import React from "react";
import { ParamGroup } from "./param-group";
import { LabeledKnob } from "./knob";
import { Select } from "./select";
import { ParamsAction } from "./state";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

export const filterDecoder = d.object({
  enabled: d.boolean,
  targetOsc: d.string(),
  kind: d.string(),
  freq: d.number(),
  q: d.number(),
  gain: d.number(),
});
type Filter = d.TypeOf<typeof filterDecoder>;

export const FilterGroup = React.memo(
  (o: { value: Filter; dispatch: React.Dispatch<ParamsAction> }) => {
    checkRenderingExclusive("params", "filter");
    const onChangeFilterEnabled = (value: boolean) => {
      o.dispatch({
        type: "changedFilterEnabled",
        value,
      });
    };
    return (
      <ParamGroup
        label="Filter"
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeFilterEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <FilterTargetOsc dispatch={o.dispatch} value={o.value.targetOsc} />
          <FilterKind dispatch={o.dispatch} value={o.value.kind} />
          <FilterFreq dispatch={o.dispatch} value={o.value.freq} />
          <FilterQ dispatch={o.dispatch} value={o.value.q} />
          <FilterGain dispatch={o.dispatch} value={o.value.gain} />
        </div>
      </ParamGroup>
    );
  }
);

const FilterTargetOsc = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const theirList = ["all", "0", "1"];
    const ourList = ["OSC 1 & 2", "OSC 1", "OSC 2"];
    const onChange = (value: string) =>
      o.dispatch({
        type: "changedFilterTargetOsc",
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

const FilterKind = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedFilterKind", value });
    return (
      <Select
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
        onChange={onChange}
      />
    );
  }
);
const FilterFreq = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFilterFreq", value });
    return (
      <LabeledKnob
        min={30}
        max={20000}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Freq"
      />
    );
  }
);
const FilterQ = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFilterQ", value });
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
const FilterGain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFilterGain", value });
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
