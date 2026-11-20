+++
title = "My Homelab"
date = "2026-08-14"
draft = false
tags = ["homelab", "kubernetes", "talos", "terraform"]
description = "Nine VMs on one Proxmox box, running a highly available Kubernetes cluster on Talos Linux."
readingTime = false
+++

I run a Kubernetes cluster in my homelab. It is nine VMs on a single Proxmox box.

Three are control plane nodes with embedded etcd. Three are workers. Three are tainted database workers. It runs on [Talos Linux](https://www.talos.dev).

Talos has no SSH and no shell. It takes its whole configuration through its own API. I like that. There is nothing to log into and fix by hand.

The full setup is public in my [infrastructure repo](https://github.com/alexraskin/infrastructure). It is not a template. You should not try to `terraform apply` it. The addresses, VM IDs, and zone IDs are mine. It is there as a worked example, and because I am quite fond of it.

## How it is built

![Homelab Diagram](/img/homelab-diagram.svg)

Terraform builds the cluster, not just the VMs. One apply creates the VMs, pushes a machine config to each node, and bootstraps etcd.

One file is the source of truth. `cluster.auto.tfvars.json` holds every address, VM ID, disk size, the VIP, and the Talos and Kubernetes versions. Terraform reads it. My scripts read the same file with `jq`. There is no second copy.

There is no golden image, no cloud-init, and no Ansible step after. The whole thing rebuilds from a `terraform apply` and a git push.

Every command is a `mise` task.

```bash
mise run bootstrap
mise run health
mise run status
```

## How it fits together

Cilium is the CNI. I install it outside Flux on purpose. Nothing that reconciles from inside the cluster can install the reason the cluster has a pod network.

Flux owns everything above that. Everything in `apps/` is deployed by Flux. Image tags are rewritten into git by image-automation, not by me.

Storage is static. Talos ships no provisioner, so most volumes are written by hand and pinned to a node.

The exception is NFS. A small Debian LXC on the same Proxmox box exports a 200 GB disk, and `csi-driver-nfs` turns it into the one StorageClass that provisions on demand. It is for anything that needs `ReadWriteMany`, or does not care which node it lands on. Terraform builds it next to the VMs, and it is in the same nightly vzdump to the NAS.

The database workers run CloudNativePG on a second disk. Backups go to object storage outside the house.

Secrets are SOPS encrypted to a single age key. The key lives in 1Password. Nothing sensitive is in git in plain text.

## Reaching it

No port is open to the internet at home.

Nothing in the cluster is public. I reach it over Tailscale. My homelab firewall has nothing to forward.

Talos cannot run `tailscaled` itself. The Tailscale operator runs in the cluster instead. A Connector re-advertises the cluster subnet, so `kubectl` works from anywhere. Individual services get their own tailnet device and a MagicDNS name.

The only public port I have is on a small NixOS VM in Oracle Cloud. HAProxy terminates TLS there and forwards over the tailnet, mostly to Plex. It is also a Tailscale exit node. Its node exporter only listens on the tailnet, so the cluster scrapes it like any other host.

![Grafana dashboard for the cloud edge VM](/img/edge.png)

Plex streams go out through that VM, so the dashboard tracks egress per month. The free tier never surprises me.

## Watching it

Monitoring is kube-prometheus-stack. Logs go to Loki, with chunks in R2.

The Proxmox host is scraped by `pve-exporter`. One box, ten guests, and every VM's CPU, memory, disk, and network on one page.

![Grafana dashboard for the Proxmox host and its guests](/img/pve.png)

Host memory sits around 80%. That is the real limit of running nine VMs on one machine, not CPU.

Postgres gets its own dashboard from the CloudNativePG metrics. Three instances, one primary and two replicas, with replication lag at zero. Base backups and WAL go to object storage, kept for 30 days.

![Grafana dashboard for the CloudNativePG cluster](/img/cnpg.png)

For the quick "is it up" answer I use [Gatus](https://gatus.io). It checks every node, the API VIP, the edge, and the platform services, and alerts me on Discord. Its history is stored in the same CNPG cluster.

![Gatus status page showing nodes and services as healthy](/img/status_page.png)

Those red bars on `cloud-edge` were me rebooting it.

## Why I keep it this way

I learn more from running this than from a cluster that hides everything. I broke most of this at least once. That is the point.

It is also useful. It hosts my own services, gives me a real place to test ideas from work, and it stays up when one node goes down.

If I need to rebuild it, I do not restore from notes. I run Terraform, install Cilium, and let Flux do the rest.
