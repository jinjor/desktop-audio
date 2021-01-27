import React from "react";

export const Radio = (o: {
  list: string[];
  value: string;
  onChange: (value: string) => void;
}) => {
  return (
    <div style={{ display: "flex", flexFlow: "column" }}>
      {o.list.map((item) => {
        const selected = item === o.value;
        return (
          <Option
            key={item}
            label={item}
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
  label: string;
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
          width: "8px",
          height: "8px",
          borderRadius: "50%",
          border: "solid 2px #222",
          backgroundColor: o.selected ? "#97f" : "#222",
          marginRight: "2px",
        }}
      />
      <div>{o.label}</div>
    </div>
  );
};
