Governance
==========

Governance related implementation including proposals and voting.

.. http:get::  /api/v1/build/{id}/

   :arg id: A Build id.

    Retrieve a single Build.

   .. sourcecode:: js

      {
          "date": "2012-03-12T19:58:29.307403",
          "error": "SPHINX ERROR",
          "id": "91207",
          "output": "SPHINX OUTPUT",
          "project": "/api/v1/project/2599/",
          "resource_uri": "/api/v1/build/91207/",
          "setup": "HEAD is now at cd00d00 Merge pull request #181 from Nagyman/solr_setup\n",
          "setup_error": "",
          "state": "finished",
          "success": true,
          "type": "html",
          "version": "/api/v1/version/37405/"
      }


   :>json string date: Date of Build.
   :>json string error: Error from Sphinx build process.
   :>json string id: Build id.
   :>json string output: Output from Sphinx build process.
   :>json string project: URI for Project of Build.
   :>json string resource_uri: URI for Build.
   :>json string setup: Setup output from Sphinx build process.
   :>json string setup_error: Setup error from Sphinx build process.
   :>json string state: "triggered", "building", or "finished"
   :>json boolean success: Was build successful?
   :>json string type: Build type ("html", "pdf", "man", or "epub")
   :>json string version: URI for Version of Build.
