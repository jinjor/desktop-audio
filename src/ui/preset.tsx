import React from "react";
import { Action } from "./state";
import * as d from "./decoder";
import { Select } from "./select";

export const presetMetaDecoder = d.object({
  name: d.string(),
});
export type PresetMeta = d.TypeOf<typeof presetMetaDecoder>;

export const Presets = React.memo(
  (o: {
    list: PresetMeta[];
    dispatch: React.Dispatch<Action>;
    name: string | null;
  }) => {
    const onChangePreset = (value: string) =>
      o.dispatch({ type: "changedPreset", value });
    return (
      <div>
        <PresetSelect
          list={o.list}
          onChange={onChangePreset}
          value={o.name ?? ""}
        />
      </div>
    );
  }
);
const PresetSelect = React.memo(
  (o: {
    list: PresetMeta[];
    onChange: (value: string) => void;
    value: string;
  }) => {
    // TODO: avoid selecting the first item
    const list = ["", ...o.list.map((item) => item.name)];
    return <Select list={list} value={o.value} onChange={o.onChange} />;
  }
);
