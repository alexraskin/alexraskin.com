+++
title = "Abfahrt Board"
date = "2026-10-15"
draft = false
tags = ["hamburg", "esp32", "trains"]
description = "I built a small departure board for my fiancée, so she does not have to check her phone to catch the train."
readingTime = false
+++

My fiancée was always complaining about the train.

Not the train itself. The checking. She would be getting ready to leave, then have to stop, open her phone, wait for the app to load, and find out if the U2 was on time, delayed, or not running at all because of a closure somewhere.

I thought that was a problem I could fix. And honestly, it sounded like a fun project.

![Abfahrt board showing U2 and bus 30 departures from Osterstraße](/img/abfahrt-board.jpeg)

I had wanted to learn more about ESPs for a while. I bought a Cheap Yellow Display, an ESP32 with a small 2.8 inch screen, and started there. With help from AI on the firmware parts I did not know, and my own infrastructure knowledge for the backend, it came together faster than I expected.

I did not want the board itself calling the transit APIs. It is a tiny device. It should not hold those credentials, work out two different data sources, or ask for the same departures again and again with no caching.

So the board does almost nothing. It polls my backend every 30 seconds, draws what it gets back, and scrolls any alerts along the bottom. That is it.

The backend is where I went overboard. It is a Go API on AWS Lambda, behind API Gateway, with DynamoDB, private S3, and Terraform to set it all up. For one small screen on a desk. I know. It is completely overkill.

But it is the part I know how to make reliable, and I wanted it to be reliable. The Lambda merges Transitous and HVV Geofox, and caches in memory so we are not hammering them on every poll. Departures are cached for 30 seconds, announcements for 5 minutes, and they can be served stale for a while if an upstream is having a bad day. If HVV fails, it falls back to Transitous. Every poll also sends a heartbeat, which Lambda writes to DynamoDB, so I can see on a little status page if the board is alive.

It also looks after itself. The first flash was over USB, but after that it updates over the air through the same API, from S3. It has two app partitions for that. If a new firmware never manages a successful poll, it rolls back to the old one and keeps going. I do not have to go and rescue it with a cable.

WiFi and the board token are set up through a portal on first boot and stored on the device. Nothing secret is compiled into the firmware.

Now there is no phone to open. She is getting ready, she looks up, and she knows. U2 in 12 minutes, no rush. Bus in 4, time to go. Train cancelled, stay a bit longer and have another coffee.

That was the whole point. Not a big system, even if the backend pretends it is one. Just one small, kind thing on the desk that makes leaving the house a little easier.
