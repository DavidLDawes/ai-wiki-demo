**wiki**

By David Dawes, August 2024. Licensed under the Apache License, version 2.0' see the LICENSE file.

Nearly trivial example of using the Anthropic API to enhance a simple wiki style web server.

Munged together samples of a simplistic file based wiki html server with Anthropic's AI, adding a simple chat request in the back end to get the definition prepopulated for you. 

The Anthropic API Key is needed for this. Rather than store  that anywehre herre in the code, I fetch the value from the environment using ANTHROPICAPIKEY as the "key" so to speak.

Pretty simplistic, but it is effectively caching. The second time you check a given topic, it gets the copy off of disk, so it doesn't use AI to re-fetch it.
