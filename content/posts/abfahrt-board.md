+++
title = "Abfahrt Board"
date = "2026-10-15"
draft = false
tags = ["hamburg", "esp32", "trains"]
description = "A small departure board for home, so leaving for the train is just a glance up."
readingTime = false
+++

I made a little departure board for home in Hamburg.

It sits on the desk and shows the next trains from Osterstraße. U2, and the buses too. Live times, not the timetable.

The idea was simple. Mornings are busy enough. Getting ready, coffee, keys, coat. I did not want leaving the house to start with unlocking a phone and opening an app.

Now it is just a glance up. Train in 12 minutes, time to finish up. Train in 4 minutes, go now. If something is delayed, the message scrolls along the bottom.

![Abfahrt board showing U2 and bus 30 departures from Osterstraße](/img/abfahrt-board.jpeg)

## How it works

The board is an ESP32-2432S028R. People call it the Cheap Yellow Display. It is a 2.8 inch ILI9341 screen with touch, on an ESP32.

It does not talk to HVV directly. It polls a small Go API I run on AWS Lambda, behind API Gateway (overkill i know). The API merges two sources. Transitous for departures, and HVV Geofox GTI in Hamburg for the same stop, including disruption alerts. The board just asks for its stop and draws what comes back.

Station and lines are config on the API side, not in the firmware. Osterstraße is set to U2, 30, and 4 by default. Change the config in Terraform and any stop Transitous covers works.

Alerts scroll through the bottom row. Tap the screen to toggle the backlight. Hold it for a second to get a device info screen.

On first boot it opens its own WiFi portal. You join the board, pick your WiFi, and enter a board token. Credentials and the token are stored on the device. They are never compiled in.

The first flash is over USB. After that it updates itself over the air from S3. If a new image never completes a successful poll, it rolls back to the previous one on its own.

Terraform provisions the Lambda, API Gateway, DynamoDB, and S3. Every command locally is a `mise` task.

It is a small thing. It just makes the flat feel calmer, and leaving feel a little less rushed.
