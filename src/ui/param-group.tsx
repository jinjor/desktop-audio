import React, { useCallback } from "react";

export const ParamGroup = (o: {
  enabled?: boolean;
  label: string;
  children: any;
  canBypass?: boolean;
  onChangeEnabled?: (value: boolean) => void;
}) => {
  const enabled = o.enabled ?? true;
  const canBypass = o.canBypass ?? false;
  const onClick = () => {
    if (canBypass) {
      o.onChangeEnabled?.(!enabled);
    }
  };
  return (
    <div
      style={{
        display: "flex",
        flexFlow: "column",
      }}
    >
      <div
        style={{
          display: "flex",
          borderBottom: "solid 1px #aaa",
          whiteSpace: "nowrap",
          alignItems: "center",
          columnGap: "4px",
        }}
        onClick={onClick}
      >
        {o.canBypass ? (
          <div
            style={{
              width: "8px",
              height: "8px",
              backgroundColor: enabled ? "rgb(153,119,255)" : "#222",
              marginTop: "-1px",
            }}
          ></div>
        ) : null}
        <label>{o.label}</label>
      </div>
      <div
        style={{
          padding: "5px 0",
          ...(enabled ? {} : { opacity: 0.2, pointerEvents: "none" }),
        }}
      >
        {o.children}
      </div>
    </div>
  );
};

export const useCallbackWithIndex = <T,>(
  index: number,
  f: (index: number, value: T) => void
) => {
  return useCallback((value: T) => f(index, value), [index, f]);
};
