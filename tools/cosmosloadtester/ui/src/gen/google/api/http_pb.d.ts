import * as jspb from 'google-protobuf'



export class Http extends jspb.Message {
  getRulesList(): Array<HttpRule>;
  setRulesList(value: Array<HttpRule>): Http;
  clearRulesList(): Http;
  addRules(value?: HttpRule, index?: number): HttpRule;

  getFullyDecodeReservedExpansion(): boolean;
  setFullyDecodeReservedExpansion(value: boolean): Http;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Http.AsObject;
  static toObject(includeInstance: boolean, msg: Http): Http.AsObject;
  static serializeBinaryToWriter(message: Http, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Http;
  static deserializeBinaryFromReader(message: Http, reader: jspb.BinaryReader): Http;
}

export namespace Http {
  export type AsObject = {
    rulesList: Array<HttpRule.AsObject>,
    fullyDecodeReservedExpansion: boolean,
  }
}

export class HttpRule extends jspb.Message {
  getSelector(): string;
  setSelector(value: string): HttpRule;

  getGet(): string;
  setGet(value: string): HttpRule;

  getPut(): string;
  setPut(value: string): HttpRule;

  getPost(): string;
  setPost(value: string): HttpRule;

  getDelete(): string;
  setDelete(value: string): HttpRule;

  getPatch(): string;
  setPatch(value: string): HttpRule;

  getCustom(): CustomHttpPattern | undefined;
  setCustom(value?: CustomHttpPattern): HttpRule;
  hasCustom(): boolean;
  clearCustom(): HttpRule;

  getBody(): string;
  setBody(value: string): HttpRule;

  getResponseBody(): string;
  setResponseBody(value: string): HttpRule;

  getAdditionalBindingsList(): Array<HttpRule>;
  setAdditionalBindingsList(value: Array<HttpRule>): HttpRule;
  clearAdditionalBindingsList(): HttpRule;
  addAdditionalBindings(value?: HttpRule, index?: number): HttpRule;

  getPatternCase(): HttpRule.PatternCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): HttpRule.AsObject;
  static toObject(includeInstance: boolean, msg: HttpRule): HttpRule.AsObject;
  static serializeBinaryToWriter(message: HttpRule, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): HttpRule;
  static deserializeBinaryFromReader(message: HttpRule, reader: jspb.BinaryReader): HttpRule;
}

export namespace HttpRule {
  export type AsObject = {
    selector: string,
    get: string,
    put: string,
    post: string,
    pb_delete: string,
    patch: string,
    custom?: CustomHttpPattern.AsObject,
    body: string,
    responseBody: string,
    additionalBindingsList: Array<HttpRule.AsObject>,
  }

  export enum PatternCase { 
    PATTERN_NOT_SET = 0,
    GET = 2,
    PUT = 3,
    POST = 4,
    DELETE = 5,
    PATCH = 6,
    CUSTOM = 8,
  }
}

export class CustomHttpPattern extends jspb.Message {
  getKind(): string;
  setKind(value: string): CustomHttpPattern;

  getPath(): string;
  setPath(value: string): CustomHttpPattern;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CustomHttpPattern.AsObject;
  static toObject(includeInstance: boolean, msg: CustomHttpPattern): CustomHttpPattern.AsObject;
  static serializeBinaryToWriter(message: CustomHttpPattern, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CustomHttpPattern;
  static deserializeBinaryFromReader(message: CustomHttpPattern, reader: jspb.BinaryReader): CustomHttpPattern;
}

export namespace CustomHttpPattern {
  export type AsObject = {
    kind: string,
    path: string,
  }
}

