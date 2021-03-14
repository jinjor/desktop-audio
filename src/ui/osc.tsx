import React from "react";
import { ParamGroup, useCallbackWithIndex } from "./param-group";
import { LabeledKnob } from "./knob";
import { ParamsAction } from "./state";
import { WaveSelect } from "./wave-select";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

export const oscDecoder = d.object({
  enabled: d.boolean,
  kind: d.string(),
  octave: d.number(),
  coarse: d.number(),
  fine: d.number(),
  level: d.number(),
});
type OSC = d.TypeOf<typeof oscDecoder>;

export const OSCGroup = React.memo(
  (o: {
    index: number;
    value: OSC;
    dispatch: React.Dispatch<ParamsAction>;
  }) => {
    checkRenderingExclusive("params", "osc");
    const onChangeEnabled = useCallbackWithIndex(
      o.index,
      (index: number, value: boolean) =>
        o.dispatch({ type: "changedOscEnabled", index, value })
    );
    const onChangeKind = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedOscKind", index, value })
    );
    const onChangeOctave = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscOctave", index, value })
    );
    const onChangeCoarse = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscCoarse", index, value })
    );
    const onChangeFine = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscFine", index, value })
    );
    const onChangeLevel = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedOscLevel", index, value })
    );
    return (
      <ParamGroup
        label={`OSC ${o.index + 1}`}
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <div style={{ textAlign: "center" }}>
            <OscKind value={o.value.kind} onChange={onChangeKind} />
          </div>
          <OSCOctave value={o.value.octave} onChange={onChangeOctave} />
          <OSCCoarse value={o.value.coarse} onChange={onChangeCoarse} />
          <OSCFine value={o.value.fine} onChange={onChangeFine} />
          <OSCLevel value={o.value.level} onChange={onChangeLevel} />
        </div>
      </ParamGroup>
    );
  }
);

const OscKind = React.memo(
  (o: { onChange: (value: string) => void; value: string }) => {
    return (
      <WaveSelect
        list={[
          "sine",
          "triangle",
          // "square",
          "square-wt",
          "pulse",
          // "saw",
          "saw-wt",
          "noise",
        ]}
        value={o.value}
        columns={3}
        onChange={o.onChange}
      />
    );
  }
);
const OSCOctave = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = -2;
    const max = 2;
    const steps = max - min + 1;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        from={0}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Octave"
      />
    );
  }
);
const OSCCoarse = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = -12;
    const max = 12;
    const steps = max - min + 1;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        exponential={false}
        value={o.value}
        from={0}
        onChange={o.onChange}
        label="Coarse"
      />
    );
  }
);
const OSCFine = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = -100;
    const max = 100;
    const steps = max - min + 1;
    const onChange = (value: number) => o.onChange(Math.round(value));
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        exponential={false}
        value={o.value}
        from={0}
        onChange={onChange}
        label="Fine"
      />
    );
  }
);
const OSCLevel = React.memo(
  (o: { onChange: (value: number) => void; value: number }) => {
    const min = 0;
    const max = 1;
    const steps = 400;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        exponential={false}
        value={o.value}
        from={0}
        onChange={o.onChange}
        label="Level"
      />
    );
  }
);
