import React from "react";
import { ParamGroup, useCallbackWithIndex } from "./param-group";
import { LabeledKnob } from "./knob";
import { Select } from "./select";
import { ParamsAction } from "./state";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

type EnvelopeDestination = {
  name: string;
  label: string;
};
const envelopeDestinations: EnvelopeDestination[] = [
  { name: "none", label: "None" },
  { name: "osc0_volume", label: "OSC 1 Volume" },
  { name: "osc1_volume", label: "OSC 2 Volume" },
  { name: "freq", label: "Freq" },
  { name: "note_filter_freq", label: "Note-Filter Freq" },
  { name: "note_filter_q", label: "Note-Filter Q" },
  { name: "note_filter_gain", label: "Note-Filter Gain" },
  { name: "filter_freq", label: "Filter Freq" },
  { name: "filter_q", label: "Filter Q" },
  { name: "filter_gain", label: "Filter Gain" },
  { name: "lfo0_amount", label: "LFO 1 Amount" },
  { name: "lfo1_amount", label: "LFO 2 Amount" },
  { name: "lfo2_amount", label: "LFO 3 Amount" },
  { name: "lfo0_freq", label: "LFO 1 Freq" },
  { name: "lfo1_freq", label: "LFO 2 Freq" },
  { name: "lfo2_freq", label: "LFO 3 Freq" },
];
const theirList = envelopeDestinations.map((e) => e.name);
const ourList = envelopeDestinations.map((e) => e.label);
export const envelopeDecoder = d.object({
  enabled: d.boolean,
  destination: d.string(),
  kind: d.string(),
  delay: d.number(),
  attack: d.number(),
  amount: d.number(),
});
type Envelope = d.TypeOf<typeof envelopeDecoder>;

export const defaultEnvelope: Envelope = {
  enabled: false,
  destination: "none",
  kind: "coming",
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
    const onChangeEnabled = useCallbackWithIndex(
      o.index,
      (index: number, value: boolean) =>
        o.dispatch({ type: "changedEnvelopeEnabled", index, value })
    );
    const onChangeDestination = useCallbackWithIndex(
      o.index,
      (index: number, value: string) => {
        o.dispatch({
          type: "changedEnvelopeDestination",
          index,
          value: theirList[ourList.indexOf(value)],
        });
      }
    );
    const onChangeKind = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedEnvelopeKind", index, value })
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
            list={ourList}
            value={ourList[theirList.indexOf(o.value.destination)]}
            onChange={onChangeDestination}
          />
          <Select
            list={["coming", "going"]}
            value={o.value.kind}
            onChange={onChangeKind}
          />
          <EnvelopeDelay value={o.value.delay} onChange={onChangeDelay} />
          <EnvelopeAttack value={o.value.attack} onChange={onChangeAttack} />
          <EnvelopeAmount value={o.value.amount} onChange={onChangeAmount} />
        </div>
      </ParamGroup>
    );
  }
);
const EnvelopeDelay = React.memo(
  (o: { value: number; onChange: (value: number) => void }) => {
    return (
      <LabeledKnob
        min={0}
        max={1000}
        steps={400}
        exponential={true}
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
        exponential={true}
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
        min={-4}
        max={4}
        from={0}
        steps={400}
        exponential={false}
        value={o.value}
        onChange={o.onChange}
        label="Freq Amount"
      />
    );
  }
);
