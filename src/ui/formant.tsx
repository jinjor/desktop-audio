import React from "react";
import { ParamGroup } from "./param-group";
import { LabeledKnob } from "./knob";
import { Radio } from "./radio";
import { ParamsAction } from "./state";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

export const formantDecoder = d.object({
  enabled: d.boolean,
  kind: d.string(),
  tone: d.number(),
  q: d.number(),
});
type Formant = d.TypeOf<typeof formantDecoder>;

export const FormantGroup = React.memo(
  (o: { value: Formant; dispatch: React.Dispatch<ParamsAction> }) => {
    checkRenderingExclusive("params", "formant");
    const onChangeFormantEnabled = (value: boolean) => {
      o.dispatch({
        type: "changedFormantEnabled",
        value,
      });
    };
    return (
      <ParamGroup
        label="Formant"
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeFormantEnabled}
      >
        <div style={{ display: "flex", gap: "12px" }}>
          <div>
            <FormantKind dispatch={o.dispatch} value={o.value.kind} />
          </div>
          <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
            <FormantTone dispatch={o.dispatch} value={o.value.tone} />
            <FormantQ dispatch={o.dispatch} value={o.value.q} />
          </div>
        </div>
      </ParamGroup>
    );
  }
);

const FormantKind = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedFormantKind", value });
    return (
      <Radio
        list={["a", "e", "i", "o", "u"]}
        value={o.value}
        onChange={onChange}
      />
    );
  }
);
const FormantTone = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFormantTone", value });
    return (
      <LabeledKnob
        min={0.5}
        max={2}
        steps={400}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Tone"
      />
    );
  }
);
const FormantQ = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedFormantQ", value });
    return (
      <LabeledKnob
        min={0}
        max={50}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Q"
      />
    );
  }
);
