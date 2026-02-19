"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";
import {
  MagnifyingGlass,
  CircleNotch,
  CheckCircle,
  XCircle,
  ArrowRight,
  FolderOpen,
  Clock,
} from "@phosphor-icons/react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { getProbes, type Probe } from "@/lib/api";

function statusVariant(status: string) {
  switch (status.toLowerCase()) {
    case "running":
      return "running" as const;
    case "complete":
    case "completed":
    case "done":
      return "complete" as const;
    case "failed":
    case "error":
      return "failed" as const;
    default:
      return "muted" as const;
  }
}

function StatusIcon({ status }: { status: string }) {
  switch (status.toLowerCase()) {
    case "running":
      return <CircleNotch size={14} className="animate-spin" />;
    case "complete":
    case "completed":
    case "done":
      return <CheckCircle size={14} weight="fill" />;
    case "failed":
    case "error":
      return <XCircle size={14} weight="fill" />;
    default:
      return <Clock size={14} />;
  }
}

function timeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: {
      staggerChildren: 0.06,
    },
  },
};

const item = {
  hidden: { opacity: 0, y: 12, scale: 0.98 },
  show: {
    opacity: 1,
    y: 0,
    scale: 1,
    transition: {
      duration: 0.25,
      ease: [0.25, 0.1, 0.25, 1] as [number, number, number, number],
    },
  },
  exit: {
    opacity: 0,
    y: -8,
    scale: 0.95,
    transition: {
      duration: 0.15,
    },
  },
};

export default function DashboardPage() {
  const [probes, setProbes] = useState<Probe[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");

  useEffect(() => {
    getProbes()
      .then((data) => {
        setProbes(data ?? []);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  const filtered = probes.filter(
    (p) =>
      p.id.toLowerCase().includes(search.toLowerCase()) ||
      p.target.toLowerCase().includes(search.toLowerCase()) ||
      p.type.toLowerCase().includes(search.toLowerCase()),
  );

  const stats = {
    total: probes.length,
    running: probes.filter((p) => p.status === "running").length,
    complete: probes.filter((p) =>
      ["complete", "completed", "done"].includes(p.status),
    ).length,
    failed: probes.filter((p) => ["failed", "error"].includes(p.status)).length,
  };

  return (
    <div className="space-y-8">
      {/* Page heading */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Dashboard
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Overview of all security probes and audit results.
        </p>
      </motion.div>

      {/* Stats row */}
      <motion.div
        className="grid grid-cols-2 gap-3 sm:grid-cols-4"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.1 }}
      >
        {[
          {
            label: "Total Probes",
            value: stats.total,
            color: "text-foreground",
          },
          { label: "Running", value: stats.running, color: "text-secondary" },
          { label: "Complete", value: stats.complete, color: "text-success" },
          { label: "Failed", value: stats.failed, color: "text-destructive" },
        ].map((stat) => (
          <Card key={stat.label} className="px-4 py-3">
            <p className="text-xs font-medium text-muted-foreground">
              {stat.label}
            </p>
            <p className={`text-2xl font-semibold tabular-nums ${stat.color}`}>
              {loading ? "—" : stat.value}
            </p>
          </Card>
        ))}
      </motion.div>

      {/* Search */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.3, delay: 0.15 }}
        className="relative"
      >
        <MagnifyingGlass
          size={16}
          className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none"
        />
        <Input
          placeholder="Search probes by ID, target, or type…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-9"
        />
      </motion.div>

      {/* Probe cards grid */}
      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader className="pb-3">
                <div className="h-4 w-32 rounded bg-muted" />
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="h-3 w-full rounded bg-muted" />
                <div className="h-3 w-2/3 rounded bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : error ? (
        <Card className="border-destructive/30 bg-destructive/5">
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <XCircle
              size={32}
              className="text-destructive mb-3"
              weight="duotone"
            />
            <p className="text-sm font-medium text-destructive">
              Failed to load probes
            </p>
            <p className="mt-1 text-xs text-muted-foreground">{error}</p>
          </CardContent>
        </Card>
      ) : filtered.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16 text-center">
            <FolderOpen
              size={40}
              className="text-muted-foreground/50 mb-3"
              weight="duotone"
            />
            <p className="text-sm font-medium text-muted-foreground">
              {search ? "No probes match your search" : "No probes yet"}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              {search
                ? "Try adjusting your search query."
                : "Run a probe from the CLI to get started."}
            </p>
          </CardContent>
        </Card>
      ) : (
        <AnimatePresence mode="popLayout">
          <motion.div
            className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
            variants={container}
            initial="hidden"
            animate="show"
          >
            {filtered.map((probe) => (
              <motion.div
                key={probe.id}
                variants={item}
                layout
                exit="exit"
              >
                <Link href={`/probes/${probe.id}`}>
                  <Card className="group relative cursor-pointer transition-all duration-200 hover:shadow-md hover:border-primary/30 hover:-translate-y-0.5">
                    <CardHeader className="pb-2">
                    <div className="flex items-start justify-between gap-2">
                      <CardTitle className="text-sm font-medium leading-snug line-clamp-2 group-hover:text-primary transition-colors">
                        {probe.id}
                      </CardTitle>
                      <Badge
                        variant={statusVariant(probe.status)}
                        className="shrink-0 gap-1"
                      >
                        <StatusIcon status={probe.status} />
                        {probe.status}
                      </Badge>
                    </div>
                  </CardHeader>

                  <CardContent className="space-y-3">
                    <div className="space-y-1.5">
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span className="font-medium text-foreground/70">
                          Target
                        </span>
                        <span className="truncate">{probe.target}</span>
                      </div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span className="font-medium text-foreground/70">
                          Type
                        </span>
                        <Badge
                          variant="outline"
                          className="text-[10px] px-1.5 py-0"
                        >
                          {probe.type}
                        </Badge>
                      </div>
                    </div>

                    <div className="flex items-center justify-between pt-1 border-t border-border/50">
                      <span className="text-[11px] text-muted-foreground flex items-center gap-1">
                        <Clock size={12} />
                        {timeAgo(probe.created_at)}
                      </span>
                      <ArrowRight
                        size={14}
                        className="text-muted-foreground/0 group-hover:text-primary transition-all duration-200 -translate-x-1 group-hover:translate-x-0 group-hover:opacity-100"
                      />
                    </div>
                  </CardContent>
                </Card>
              </Link>
            </motion.div>
          ))}
        </motion.div>
      </AnimatePresence>
      )}
    </div>
  );
}
