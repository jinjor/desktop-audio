import React from "react";

export const Radio = (o: {
  list: string[];
  value: string;
  columns?: number;
  toElement?: (value: string) => string | JSX.Element;
  onChange: (value: string) => void;
}) => {
  const columns = o.columns || 1;
  return (
    // <div style={{ display: "flex", flexFlow: "column", userSelect: "none" }}>
    <div
      style={{
        display: "grid",
        gridTemplateColumns: `repeat(${columns},1fr)`,
        gap: "2px 6px",
      }}
    >
      {o.list.map((item) => {
        const selected = item === o.value;
        return (
          <Option
            key={item}
            label={o.toElement?.(item) ?? item}
            selected={selected}
            onClick={() => {
              if (!selected) {
                o.onChange(item);
              }
            }}
          />
        );
      })}
    </div>
  );
};

const Option = (o: {
  label: string | JSX.Element;
  selected: boolean;
  onClick: () => void;
}) => {
  return (
    <div
      style={{ display: "inline-flex", alignItems: "center" }}
      onClick={() => o.onClick()}
    >
      <div
        style={{
          width: "7px",
          height: "7px",
          borderRadius: "50%",
          border: "solid 2px #222",
          backgroundColor: o.selected ? "rgb(153,119,255)" : "#000",
          marginRight: "2px",
        }}
      />
      <div>{o.label}</div>
    </div>
  );
};
