declare module 'crypto-js' {
  interface WordArray {
    words: number[];
    sigBytes: number;
    toString(encoder?: any): string;
    concat(wordArray: WordArray): WordArray;
    clamp(): void;
    clone(): WordArray;
  }

  interface Encoder {
    stringify(wordArray: WordArray): string;
    parse(str: string): WordArray;
  }

  interface HashAlgorithm {
    reset(): HashAlgorithm;
    update(message: string | WordArray): HashAlgorithm;
    finalize(message?: string | WordArray): WordArray;
  }

  interface CipherParams {
    ciphertext: WordArray;
    key: WordArray;
    iv: WordArray;
    salt: WordArray;
    algorithm: any;
    mode: any;
    padding: any;
    blockSize: number;
    formatter: any;
  }

  interface CipherOption {
    iv?: string | WordArray;
    mode?: any;
    padding?: any;
    format?: any;
  }

  export function MD5(message: string | WordArray): WordArray;
  export function SHA1(message: string | WordArray): WordArray;
  export function SHA256(message: string | WordArray): WordArray;
  export function SHA224(message: string | WordArray): WordArray;
  export function SHA512(message: string | WordArray): WordArray;
  export function SHA384(message: string | WordArray): WordArray;
  export function SHA3(message: string | WordArray, outputLength?: number): WordArray;
  export function RIPEMD160(message: string | WordArray): WordArray;
  export function HmacMD5(message: string | WordArray, key: string | WordArray): WordArray;
  export function HmacSHA1(message: string | WordArray, key: string | WordArray): WordArray;
  export function HmacSHA256(message: string | WordArray, key: string | WordArray): WordArray;
  export function HmacSHA224(message: string | WordArray, key: string | WordArray): WordArray;
  export function HmacSHA512(message: string | WordArray, key: string | WordArray): WordArray;
  export function HmacSHA384(message: string | WordArray, key: string | WordArray): WordArray;
  export function HmacSHA3(message: string | WordArray, key: string | WordArray, outputLength?: number): WordArray;
  export function HmacRIPEMD160(message: string | WordArray, key: string | WordArray): WordArray;
  export function PBKDF2(password: string | WordArray, salt: string | WordArray, cfg?: any): WordArray;
  export function AES(message: string | WordArray, key: string | WordArray, cfg?: CipherOption): WordArray;
  export function TripleDES(message: string | WordArray, key: string | WordArray, cfg?: CipherOption): WordArray;
  export function RC4(message: string | WordArray, key: string | WordArray, cfg?: CipherOption): WordArray;
  export function Rabbit(message: string | WordArray, key: string | WordArray, cfg?: CipherOption): WordArray;

  // 添加更多需要的类型声明...
} 