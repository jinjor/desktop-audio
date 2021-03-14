import React from "react";
import { ParamGroup } from "./param-group";
import { LabeledKnob } from "./knob";
import { ParamsAction } from "./state";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

export const echoDecoder = d.object({
  enabled: d.boolean,
  delay: d.number(),
  feedbackGain: d.number(),
  mix: d.number(),
});
type Echo = d.TypeOf<typeof echoDecoder>;

export const EchoGroup = React.memo(
  (o: { value: Echo; dispatch: React.Dispatch<ParamsAction> }) => {
    checkRenderingExclusive("params", "echo");
    const onChangeEchoEnabled = (value: boolean) => {
      o.dispatch({
        type: "changedEchoEnabled",
        value,
      });
    };
    return (
      <ParamGroup
        label="Echo"
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeEchoEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <EchoDelay value={o.value.delay} dispatch={o.dispatch} />
          <EchoFeedbackGain
            value={o.value.feedbackGain}
            dispatch={o.dispatch}
          />
          <EchoMix value={o.value.mix} dispatch={o.dispatch} />
        </div>
      </ParamGroup>
    );
  }
);

const EchoDelay = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedEchoDelay", value });
    return (
      <LabeledKnob
        min={10}
        max={800}
        steps={800}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Delay"
      />
    );
  }
);
const EchoFeedbackGain = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedEchoFeedbackGain", value });
    return (
      <LabeledKnob
        min={0}
        max={0.5}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="FeedbackGain"
      />
    );
  }
);
const EchoMix = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedEchoMix", value });
    return (
      <LabeledKnob
        min={0}
        max={1.0}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={onChange}
        label="Mix"
      />
    );
  }
);
