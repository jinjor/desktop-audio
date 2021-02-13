export interface Decoder<T> {
  run(value: unknown): T;
}

export class DecodeError extends Error {}

export const any: Decoder<unknown> = {
  run(value: unknown) {
    return value;
  },
};

export function is<T>(v: T): Decoder<T> {
  return {
    run(value: unknown) {
      if (v === value) {
        return value as T;
      }
      throw new DecodeError(value + " is not " + v + "!");
    },
  };
}

export const boolean: Decoder<boolean> = {
  run(value: unknown) {
    if (typeof value !== "boolean") {
      throw new DecodeError(value + " is not a boolean!");
    }
    return value;
  },
};

type IntOptions = {
  min?: number;
  max?: number;
};
type NumberOptions = IntOptions & {
  isInt?: boolean;
};

export function number(options?: NumberOptions): Decoder<number> {
  const { min, max, isInt = false } = options ?? {};
  return {
    run(value: unknown) {
      if (typeof value !== "number") {
        throw new DecodeError(value + " is not a nubmer!");
      }
      if (isInt && !Number.isInteger(value)) {
        throw new DecodeError(value + " is not an integer!");
      }
      if (min != null && value < min) {
        throw new DecodeError(value + " is less than " + min);
      }
      if (max != null && value > max) {
        throw new DecodeError(value + " is greater than " + max);
      }
      return value;
    },
  };
}

export function int(options?: IntOptions): Decoder<number> {
  return number({ isInt: true, ...options });
}

export function string(options?: {
  minLength?: number;
  maxLength?: number;
}): Decoder<string> {
  const { minLength, maxLength } = options ?? {};
  return {
    run(value: unknown) {
      if (typeof value !== "string") {
        throw new DecodeError(value + " is not a string!");
      }
      if (minLength != null && value.length < minLength) {
        throw new DecodeError(value + " is longer than " + minLength);
      }
      if (maxLength != null && value.length > maxLength) {
        throw new DecodeError(value + " is shorter than " + maxLength);
      }
      return value;
    },
  };
}

export function pattern(regex: RegExp): Decoder<string> {
  return {
    run(value: unknown) {
      const s = string().run(value);
      if (!s.match(regex)) {
        throw new DecodeError(value + " does not match the pattern " + regex);
      }
      return s;
    },
  };
}

export function optional<T>(
  d: Decoder<T>,
  defaultValue?: T
): Decoder<T | undefined> {
  return {
    run(value: unknown) {
      if (value == null) {
        return defaultValue;
      }
      return d.run(value);
    },
  };
}

export function array<T>(d: Decoder<T>): Decoder<T[]> {
  return {
    run(value: unknown) {
      if (!Array.isArray(value)) {
        throw new DecodeError(value + " is not an array!");
      }
      return value.map(d.run.bind(d));
    },
  };
}

export function object<T>(d: { [K in keyof T]: Decoder<T[K]> }): Decoder<T> {
  return {
    run(value: unknown): T {
      if (typeof value !== "object" || value === null || Array.isArray(value)) {
        throw new DecodeError(value + " is not an object!");
      }
      const ret: any = {};
      for (const key in d) {
        ret[key] = d[key].run((value as any)[key]);
      }
      return ret;
    },
  };
}

export function dict<V>(d: Decoder<V>): Decoder<{ [key: string]: V }> {
  return {
    run(value: unknown): { [key: string]: V } {
      if (typeof value !== "object" || value === null || Array.isArray(value)) {
        throw new DecodeError(value + " is not an object!");
      }
      const ret: { [key: string]: V } = {};
      for (const key in value) {
        ret[key] = d.run((value as any)[key]);
      }
      return ret;
    },
  };
}

export function oneOf<T>(d: Decoder<T>[]): Decoder<T> {
  return {
    run(value: unknown): T {
      for (const decoder of d) {
        try {
          return decoder.run(value);
        } catch (e) {}
      }
      throw new DecodeError(
        value + " cannot be decoded by any of " + d.length + " decoders!"
      );
    },
  };
}

export function keywords<T, V extends T>(keywords: V[]): Decoder<T> {
  return {
    run(value: unknown): T {
      for (const keyword of keywords) {
        if (keyword === value) {
          return keyword;
        }
      }
      throw new DecodeError(value + " should be one of " + keywords);
    },
  };
}

export function map<T, U>(f: (t: T) => U, d: Decoder<T>): Decoder<U> {
  return {
    run(value: unknown): U {
      return f(d.run(value));
    },
  };
}

function strAs<T>(convert: (value: string) => T): Decoder<T> {
  return {
    run(value: unknown): T {
      const s = string().run(value);
      return convert(s);
    },
  };
}

export function strAsNumber(options?: NumberOptions): Decoder<number> {
  return strAs((s) => {
    const n = parseFloat(s);
    return number(options).run(n);
  });
}

export function strAsInt(options?: IntOptions): Decoder<number> {
  return strAsNumber({ isInt: true, ...options });
}

export function strAsBoolean(): Decoder<boolean> {
  return strAs((s) => {
    if (s === "true") {
      return true;
    } else if (s === "false") {
      return false;
    }
    throw new DecodeError(s + " is not a boolean!");
  });
}
