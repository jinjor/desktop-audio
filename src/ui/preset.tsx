import React from "react";
import { Action } from "./state";
import * as d from "./decoder";
import { Select } from "./select";
import { ScheduleFn } from "./react-util";

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
      <PresetSelect
        list={o.list}
        onChange={onChangePreset}
        value={o.name ?? ""}
      />
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

export type PresetSaverAction =
  | { type: "inputName"; value: string }
  | { type: "save" }
  | { type: "close" };
export const presetSaverReducer = (
  state: PresetSaverState,
  action: PresetSaverAction,
  schedule: ScheduleFn<Action>
): PresetSaverState => {
  switch (action.type) {
    case "inputName":
      return { ...state, name: action.value };
    case "save":
      schedule((dispatch) =>
        dispatch({ type: "savePreset", value: state.name })
      );
      return { ...state, open: false };
    case "close":
      return { ...state, open: false };
  }
};
export type PresetSaverState = {
  open: boolean;
  name: string;
};
export const initialPresetSaverState: PresetSaverState = {
  open: false,
  name: "",
};
export const openPresetSaver = (name: string): PresetSaverState => ({
  open: true,
  name,
});
export const PresetSaver = React.memo(
  (o: {
    state: PresetSaverState;
    dispatch: React.Dispatch<PresetSaverAction>;
  }) => {
    const onChangeName = (e: React.ChangeEvent<HTMLInputElement>) => {
      o.dispatch({ type: "inputName", value: e.target.value });
    };
    const onClickOK = () => {
      o.dispatch({ type: "save" });
    };
    const onClickCancel = () => {
      o.dispatch({ type: "close" });
    };
    const onClickClose = () => {
      o.dispatch({ type: "close" });
    };
    if (!o.state.open) {
      return null;
    }
    return (
      <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0 }}>
        <div>
          <input onChange={onChangeName} value={o.state.name} />
          <button onClick={onClickOK}>OK</button>
          <button onClick={onClickCancel}>Cancel</button>
          <button onClick={onClickClose}>X</button>
        </div>
        <div style={{ backgroundColor: "rgba(0,0,0,0.5)" }}></div>
      </div>
    );
  }
);
