import React from "react";
import { ParamGroup, useCallbackWithIndex } from "./param-group";
import { LabeledKnob } from "./knob";
import { Select } from "./select";
import { ParamsAction } from "./state";
import { WaveSelect } from "./wave-select";
import * as d from "./decoder";
import { checkRenderingExclusive } from "./debug";

type LFODestination = {
  freqType: string;
  minFreq: number;
  maxFreq: number;
  defaultFreq: number;
  minAmount: number;
  maxAmount: number;
  defaultAmount: number;
  fromAmount: number;
};
export const lfoDestinations = new Map<string, LFODestination>();
lfoDestinations.set("none", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 0,
  defaultFreq: 0,
  minAmount: 0,
  maxAmount: 0,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("tremolo", {
  freqType: "absolute",
  minFreq: 0.1,
  maxFreq: 100,
  defaultFreq: 0.1,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("vibrato", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 40,
  defaultFreq: 10,
  minAmount: 0,
  maxAmount: 200,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("fm", {
  freqType: "ratio",
  minFreq: 0.1,
  maxFreq: 10,
  defaultFreq: 3,
  minAmount: -2400,
  maxAmount: 2400,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("pm", {
  freqType: "ratio",
  minFreq: 0.1,
  maxFreq: 10,
  defaultFreq: 3,
  minAmount: 0,
  maxAmount: 1.57,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("am", {
  freqType: "ratio",
  minFreq: 0.1,
  maxFreq: 10,
  defaultFreq: 3,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("note_filter_freq", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 40,
  defaultFreq: 10,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});
lfoDestinations.set("filter_freq", {
  freqType: "absolute",
  minFreq: 0,
  maxFreq: 40,
  defaultFreq: 10,
  minAmount: 0,
  maxAmount: 1,
  defaultAmount: 0,
  fromAmount: 0,
});

export const lfoDecoder = d.object({
  enabled: d.boolean,
  destination: d.string(),
  wave: d.string(),
  freq: d.number(),
  amount: d.number(),
});
type LFO = d.TypeOf<typeof lfoDecoder>;

export const defaultLFO = {
  enabled: false,
  destination: "none",
  wave: "sine",
  freq: 0,
  amount: 0,
};

export const LFOGroup = React.memo(
  (o: {
    index: number;
    value: LFO;
    dispatch: React.Dispatch<ParamsAction>;
  }) => {
    checkRenderingExclusive("params", "lfo");
    const list = [...lfoDestinations.keys()];
    const destination = lfoDestinations.get(o.value.destination)!;
    const onChangeEnabled = useCallbackWithIndex(
      o.index,
      (index: number, value: boolean) =>
        o.dispatch({ type: "changedLFOEnabled", index, value })
    );
    const onChangeDestination = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedLFODestination", index, value })
    );
    const onChangeWave = useCallbackWithIndex(
      o.index,
      (index: number, value: string) =>
        o.dispatch({ type: "changedLFOWave", index, value })
    );
    const onChangeFreq = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedLFOFreq", index, value })
    );
    const onChangeAmount = useCallbackWithIndex(
      o.index,
      (index: number, value: number) =>
        o.dispatch({ type: "changedLFOAmount", index, value })
    );
    return (
      <ParamGroup
        label={`LFO ${o.index + 1}`}
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
          <div style={{ textAlign: "center" }}>
            <LFOWave value={o.value.wave} onChange={onChangeWave} />
          </div>
          <LFOFreq
            min={destination.minFreq}
            max={destination.maxFreq}
            value={o.value.freq}
            onChange={onChangeFreq}
          />
          <LFOAmount
            min={destination.minAmount}
            max={destination.maxAmount}
            value={o.value.amount}
            from={destination.fromAmount}
            onChange={onChangeAmount}
          />
        </div>
      </ParamGroup>
    );
  }
);
const LFOWave = React.memo(
  (o: { value: string; onChange: (value: string) => void }) => {
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
          "saw-rev",
        ]}
        value={o.value}
        columns={3}
        onChange={o.onChange}
      />
    );
  }
);
const LFOFreq = React.memo(
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
        label="Freq"
      />
    );
  }
);
const LFOAmount = React.memo(
  (o: {
    min: number;
    max: number;
    value: number;
    from: number;
    onChange: (value: number) => void;
  }) => {
    return (
      <LabeledKnob
        min={o.min}
        max={o.max}
        steps={400}
        exponential={false}
        value={o.value}
        from={o.from}
        onChange={o.onChange}
        label="Amount"
      />
    );
  }
);
