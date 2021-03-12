import { useEffect, useReducer, useState } from "react";

export type Schedule<A> = (d: React.Dispatch<A>) => void;
export type ScheduleFn<A> = (s: Schedule<A>) => void;
export type ReducerWithEffect<S, A> = (
  state: S,
  action: A,
  schedule: ScheduleFn<A>
) => S;
export const useReducerWithEffect = <S, A>(
  reducer: ReducerWithEffect<S, A>,
  initialState: S
): [S, React.Dispatch<A>] => {
  const [effect, setEffect] = useState({ f: () => {} });
  useEffect(() => {
    effect.f();
  }, [effect]);
  const [state, dispatch] = useReducer((state: S, action: A) => {
    const schedules: Schedule<A>[] = [];
    const newState = reducer(state, action, schedules.push.bind(schedules));
    setEffect({
      f: () => {
        for (const schedule of schedules) {
          schedule(dispatch);
        }
      },
    });
    return newState;
  }, initialState);
  return [state, dispatch];
};
