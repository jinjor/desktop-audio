import React from "react";
import { ParamGroup } from "./param-group";
import { LabeledKnob } from "./knob";
import { Radio } from "./radio";
import { ParamsAction } from "./state";
import { checkRenderingExclusive } from "./debug";

export const ModeGroup = React.memo(
  (o: {
    poly: string;
    glideTime: number;
    velSense: number;
    dispatch: React.Dispatch<ParamsAction>;
  }) => {
    checkRenderingExclusive("params", "mode");
    return (
      <ParamGroup label="NOTE">
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <Poly value={o.poly} dispatch={o.dispatch} />
          <GlideTime value={o.glideTime} dispatch={o.dispatch} />
          <VelSense value={o.velSense} dispatch={o.dispatch} />
        </div>
      </ParamGroup>
    );
  }
);

const Poly = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: string }) => {
    const onChange = (value: string) =>
      o.dispatch({ type: "changedPoly", value });
    return (
      <Radio list={["mono", "poly"]} value={o.value} onChange={onChange} />
    );
  }
);
const GlideTime = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedGlideTime", value });
    const min = 1;
    const max = 400;
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
        label="Glide"
      />
    );
  }
);
const VelSense = React.memo(
  (o: { dispatch: React.Dispatch<ParamsAction>; value: number }) => {
    const onChange = (value: number) =>
      o.dispatch({ type: "changedVelSense", value });
    const min = 0;
    const max = 1;
    const steps = 400;
    return (
      <LabeledKnob
        min={min}
        max={max}
        steps={steps}
        from={0}
        exponential={true}
        value={o.value}
        onChange={onChange}
        label="Sense"
      />
    );
  }
);
