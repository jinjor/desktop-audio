import { LabeledKnob } from "./knob";
import { ParamsAction } from "./state";
import React from "react";
import { ParamGroup } from "./param-group";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

export const adsrDecoder = d.object({
  attack: d.number(),
  decay: d.number(),
  sustain: d.number(),
  release: d.number(),
});
type ADSR = d.TypeOf<typeof adsrDecoder>;

export const ADSRGroup = React.memo(
  (o: { value: ADSR; dispatch: React.Dispatch<ParamsAction> }) => {
    checkRenderingExclusive("params", "adsr");
    return (
      <ParamGroup label="Envelope">
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <Attack value={o.value.attack} dispatch={o.dispatch} />
          <Decay value={o.value.decay} dispatch={o.dispatch} />
          <Sustain value={o.value.sustain} dispatch={o.dispatch} />
          <Release value={o.value.release} dispatch={o.dispatch} />
        </div>
      </ParamGroup>
    );
  }
);

const Attack = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrAttack", value });
    return (
      <LabeledKnob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Attack"
      />
    );
  }
);
const Decay = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrDecay", value });
    return (
      <LabeledKnob
        min={0}
        max={400}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Decay"
      />
    );
  }
);
const Sustain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrSustain", value });
    return (
      <LabeledKnob
        min={0}
        max={1}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Sustain"
      />
    );
  }
);
const Release = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedAdsrRelease", value });
    return (
      <LabeledKnob
        min={0}
        max={800}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Release"
      />
    );
  }
);
