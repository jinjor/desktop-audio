import React from "react";
import { ParamGroup, useCallbackWithIndex } from "./param-group";
import { LabeledKnob } from "./knob";
import { Select } from "./select";
import { ParamsAction } from "./state";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

type EnvelopeDestination = {
  minDelay: number;
  maxDelay: number;
  defaultDelay: number;
};
const envelopeDestinations = new Map<string, EnvelopeDestination>();
envelopeDestinations.set("none", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("osc0_volume", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("osc1_volume", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("freq", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("note_filter_freq", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("note_filter_q", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("note_filter_q_0v", {
  minDelay: 0,
  maxDelay: 1000,
  defaultDelay: 0,
});
envelopeDestinations.set("note_filter_gain", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("note_filter_gain_0v", {
  minDelay: 0,
  maxDelay: 1000,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_freq", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_q", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_q_0v", {
  minDelay: 0,
  maxDelay: 1000,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_gain", {
  minDelay: 0,
  maxDelay: 0,
  defaultDelay: 0,
});
envelopeDestinations.set("filter_gain_0v", {
  minDelay: 0,
  maxDelay: 1000,
  defaultDelay: 0,
});
for (let i = 0; i < 3; i++) {
  envelopeDestinations.set(`lfo${i}_amount`, {
    minDelay: 0,
    maxDelay: 1000,
    defaultDelay: 200,
  });
}
for (let i = 0; i < 3; i++) {
  envelopeDestinations.set(`lfo${i}_freq`, {
    minDelay: 0,
    maxDelay: 0,
    defaultDelay: 0,
  });
}

export const envelopeDecoder = d.object({
  enabled: d.boolean,
  destination: d.string(),
  delay: d.number(),
  attack: d.number(),
  amount: d.number(),
});
type Envelope = d.TypeOf<typeof envelopeDecoder>;

export const defaultEnvelope: Envelope = {
  enabled: false,
  destination: "none",
  delay: 0,
  attack: 0,
  amount: 0,
};
export const EnvelopeGroup = React.memo(
  (o: {
    index: number;
    value: Envelope;
    dispatch: React.Dispatch<ParamsAction>;
  }) => {
    checkRenderingExclusive("params", "envelope");
    const list = [...envelopeDestinations.keys()];
    const destination = envelopeDestinations.get(o.value.destination)!;
    const onChangeEnabled = useCallbackWithIndex(
      o.index,
      (index: number, value: boolean) =>
        o.dispatch({ type: "changedEnvelopeEnabled", index, value })
    );
    const onChangeDestination = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedEnvelopeDestination", index, value })
    );
    const onChangeDelay = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedEnvelopeDelay", index, value })
    );
    const onChangeAttack = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedEnvelopeAttack", index, value })
    );
    const onChangeAmount = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedEnvelopeAmount", index, value })
    );
    return (
      <ParamGroup
        label={`Envelope ${o.index + 1}`}
        enabled={o.value.enabled}
        canBypass={true}
        onChangeEnabled={onChangeEnabled}
      >
        <div style={{ display: "flex", flexFlow: "column", gap: "6px" }}>
          <Select
            list={list}
            value={o.value.destination}
            onChange={onChangeDestination}
          />
          <EnvelopeDelay
            min={destination.minDelay}
            max={destination.maxDelay}
            value={o.value.delay}
            onChange={onChangeDelay}
          />
          <EnvelopeAttack value={o.value.attack} onChange={onChangeAttack} />
          <EnvelopeAmount value={o.value.amount} onChange={onChangeAmount} />
        </div>
      </ParamGroup>
    );
  }
);
const EnvelopeDelay = React.memo(
  (o: {
    min: number;
    max: number;
    value: number;
    onChange: (value: number) => void;
  }) => {
    return (
      <LabeledKnob
        min={o.min}
        max={o.max}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Delay"
      />
    );
  }
);
const EnvelopeAttack = React.memo(
  (o: { value: number; onChange: (value: number) => void }) => {
    return (
      <LabeledKnob
        min={0}
        max={1000}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Attack"
      />
    );
  }
);
const EnvelopeAmount = React.memo(
  (o: { value: number; onChange: (value: number) => void }) => {
    return (
      <LabeledKnob
        min={-1}
        max={1}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Amount"
      />
    );
  }
);
