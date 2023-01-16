// source: google/api/annotations.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {missingRequire} reports error on implicit type usages.
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!
/* eslint-disable */
// @ts-nocheck

var jspb = require('google-protobuf');
var goog = jspb;
var global = (function() {
  if (this) { return this; }
  if (typeof window !== 'undefined') { return window; }
  if (typeof global !== 'undefined') { return global; }
  if (typeof self !== 'undefined') { return self; }
  return Function('return this')();
}.call(null));

var google_api_http_pb = require('../../google/api/http_pb.js');
goog.object.extend(proto, google_api_http_pb);
var google_protobuf_descriptor_pb = require('google-protobuf/google/protobuf/descriptor_pb.js');
goog.object.extend(proto, google_protobuf_descriptor_pb);
goog.exportSymbol('proto.google.api.http', null, global);

/**
 * A tuple of {field number, class constructor} for the extension
 * field named `http`.
 * @type {!jspb.ExtensionFieldInfo<!proto.google.api.HttpRule>}
 */
proto.google.api.http = new jspb.ExtensionFieldInfo(
    72295728,
    {http: 0},
    google_api_http_pb.HttpRule,
     /** @type {?function((boolean|undefined),!jspb.Message=): !Object} */ (
         google_api_http_pb.HttpRule.toObject),
    0);

google_protobuf_descriptor_pb.MethodOptions.extensionsBinary[72295728] = new jspb.ExtensionFieldBinaryInfo(
    proto.google.api.http,
    jspb.BinaryReader.prototype.readMessage,
    jspb.BinaryWriter.prototype.writeMessage,
    google_api_http_pb.HttpRule.serializeBinaryToWriter,
    google_api_http_pb.HttpRule.deserializeBinaryFromReader,
    false);
// This registers the extension field with the extended class, so that
// toObject() will function correctly.
google_protobuf_descriptor_pb.MethodOptions.extensions[72295728] = proto.google.api.http;

goog.object.extend(exports, proto.google.api);
