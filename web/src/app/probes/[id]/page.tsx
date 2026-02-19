"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import ReactMarkdown from "react-markdown";
import rehypeRaw from "rehype-raw";
import {
  ArrowLeft,
  CheckCircle,
  Circle,
  Clock,
  FileText,
  Warning,
  WarningCircle,
  Info,
  ShieldWarning,
  Spinner,
  Trash,
} from "@phosphor-icons/react";

import { getProbe, deleteFinding as deleteFindingAPI, type Probe } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";

interface Finding {
  id: string;
  text: string;
  severity: string;
  completed: boolean;
}

function severityIcon(severity: string) {
  switch (severity.toLowerCase()) {
    case "critical":
      return (
        <ShieldWarning size={16} weight="fill" className="text-destructive" />
      );
    case "high":
      return <WarningCircle size={16} weight="fill" className="text-primary" />;
    case "medium":
      return <Warning size={16} weight="fill" className="text-warning" />;
    case "low":
      return <Info size={16} weight="fill" className="text-secondary" />;
    default:
      return <Info size={16} weight="fill" className="text-muted-foreground" />;
  }
}

function severityVariant(
  severity: string,
): "critical" | "high" | "medium" | "low" | "info" {
  const map: Record<string, "critical" | "high" | "medium" | "low" | "info"> = {
    critical: "critical",
    high: "high",
    medium: "medium",
    low: "low",
    info: "info",
  };
  return map[severity.toLowerCase()] ?? "info";
}

function statusVariant(
  status: string,
): "running" | "complete" | "failed" | "muted" {
  const map: Record<string, "running" | "complete" | "failed"> = {
    running: "running",
    complete: "complete",
    completed: "complete",
    done: "complete",
    failed: "failed",
    error: "failed",
  };
  return map[status.toLowerCase()] ?? "muted";
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

export default function ProbeDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [probe, setProbe] = useState<Probe | null>(null);
  const [findings, setFindings] = useState<Finding[]>([]);
  const [content, setContent] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"report" | "findings">("report");

  useEffect(() => {
    if (!id) return;

    async function load() {
      try {
        setLoading(true);
        const data = await getProbe(id);
        setProbe(data);

        const probeData = data as Probe & {
          content?: string;
          findings?: Finding[];
        };

        if (probeData.content) {
          setContent(probeData.content);
        } else if (probeData.file_path) {
          try {
            const res = await fetch(`/api/probes/${id}/content`);
            if (res.ok) {
              const text = await res.text();
              setContent(text);
            }
          } catch {
            // Content not available
          }
        }

        if (probeData.findings) {
          setFindings(probeData.findings);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load probe");
      } finally {
        setLoading(false);
      }
    }

    load();
  }, [id]);

  const toggleFinding = async (findingId: string) => {
    setFindings((prev) =>
      prev.map((f) =>
        f.id === findingId ? { ...f, completed: !f.completed } : f,
      ),
    );

    try {
      await fetch(`/api/findings/${findingId}`, { method: "PATCH" });
    } catch {
      setFindings((prev) =>
        prev.map((f) =>
          f.id === findingId ? { ...f, completed: !f.completed } : f,
        ),
      );
    }
  };

  const handleDeleteFinding = async (findingId: string) => {
    const previous = findings;
    setFindings((prev) => prev.filter((f) => f.id !== findingId));

    try {
      await deleteFindingAPI(findingId);
    } catch {
      setFindings(previous);
    }
  };

  const completedCount = findings.filter((f) => f.completed).length;

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="flex flex-col items-center gap-4 text-muted-foreground"
        >
          <Spinner size={32} className="animate-spin" />
          <p className="text-sm font-medium">Loading probe...</p>
        </motion.div>
      </div>
    );
  }

  if (error || !probe) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex flex-col items-center gap-4"
        >
          <WarningCircle
            size={48}
            weight="duotone"
            className="text-destructive"
          />
          <p className="text-sm text-muted-foreground">
            {error || "Probe not found"}
          </p>
          <Button variant="outline" size="sm" onClick={() => router.push("/")}>
            <ArrowLeft size={16} />
            Back to Dashboard
          </Button>
        </motion.div>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="space-y-6"
    >
      {/* Back button + header */}
      <div className="space-y-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.push("/")}
          className="gap-1.5 text-muted-foreground hover:text-foreground -ml-2"
        >
          <ArrowLeft size={16} />
          Back
        </Button>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1.5">
            <h1 className="text-2xl font-semibold tracking-tight text-foreground">
              {probe.id}
            </h1>
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
              <span className="flex items-center gap-1.5">
                <FileText size={14} />
                {probe.type}
              </span>
              <span className="flex items-center gap-1.5">
                <Clock size={14} />
                {timeAgo(probe.created_at)}
              </span>
              <span className="text-xs font-mono text-muted-foreground/70 truncate max-w-[200px]">
                {probe.target}
              </span>
            </div>
          </div>

          <Badge
            variant={statusVariant(probe.status)}
            className="self-start capitalize"
          >
            {probe.status}
          </Badge>
        </div>
      </div>

      <Separator />

      {/* Tab switcher */}
      <div className="flex gap-1 p-1 bg-muted/50 rounded-lg w-fit">
        <button
          onClick={() => setActiveTab("report")}
          className={cn(
            "px-4 py-1.5 text-sm font-medium rounded-md transition-all duration-200",
            activeTab === "report"
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground",
          )}
        >
          Report
        </button>
        <button
          onClick={() => setActiveTab("findings")}
          className={cn(
            "px-4 py-1.5 text-sm font-medium rounded-md transition-all duration-200 flex items-center gap-2",
            activeTab === "findings"
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground",
          )}
        >
          Findings
          {findings.length > 0 && (
            <span className="text-xs bg-primary/10 text-primary px-1.5 py-0.5 rounded-full font-mono">
              {completedCount}/{findings.length}
            </span>
          )}
        </button>
      </div>

      {/* Content area */}
      {activeTab === "report" ? (
        <motion.div
          key="report"
          initial={{ opacity: 0, x: -8 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.2 }}
        >
          <Card>
            <CardContent className="p-6">
              {content ? (
                <ScrollArea className="max-h-[70vh]">
                  <article className="prose prose-sm dark:prose-invert max-w-none prose-headings:text-foreground prose-p:text-muted-foreground prose-a:text-primary prose-strong:text-foreground prose-code:text-primary prose-code:bg-muted prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:text-xs prose-code:font-mono prose-pre:bg-muted prose-pre:border prose-pre:border-border prose-li:text-muted-foreground">
                    <ReactMarkdown rehypePlugins={[rehypeRaw]}>
                      {content}
                    </ReactMarkdown>
                  </article>
                </ScrollArea>
              ) : (
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <FileText
                    size={40}
                    weight="duotone"
                    className="mb-3 opacity-40"
                  />
                  <p className="text-sm font-medium">
                    No report content available
                  </p>
                  <p className="text-xs mt-1 text-muted-foreground/70">
                    The probe report will appear here once generated
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>
      ) : (
        <motion.div
          key="findings"
          initial={{ opacity: 0, x: 8 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.2 }}
          className="space-y-3"
        >
          {findings.length > 0 && (
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">
                {completedCount} of {findings.length} completed
              </span>
              <div className="h-1.5 w-32 rounded-full bg-muted overflow-hidden">
                <motion.div
                  className="h-full rounded-full bg-success"
                  initial={{ width: 0 }}
                  animate={{
                    width: `${
                      findings.length > 0
                        ? (completedCount / findings.length) * 100
                        : 0
                    }%`,
                  }}
                  transition={{ duration: 0.4 }}
                />
              </div>
            </div>
          )}

          {findings.length > 0 ? (
            <Card>
              <CardContent className="p-0">
                <ScrollArea className="max-h-[65vh]">
                  <div className="divide-y divide-border">
                    {findings.map((finding, index) => (
                      <motion.div
                        key={finding.id}
                        initial={{ opacity: 0, y: 6 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        transition={{
                          duration: 0.2,
                          delay: index * 0.03,
                        }}
                        className={cn(
                          "flex items-start gap-3 w-full px-4 py-3 text-left transition-all duration-200 hover:bg-muted/30 group",
                          finding.completed && "opacity-60",
                        )}
                      >
                        <button
                          onClick={() => toggleFinding(finding.id)}
                          className="mt-0.5 shrink-0 cursor-pointer"
                        >
                          {finding.completed ? (
                            <CheckCircle
                              size={20}
                              weight="fill"
                              className="text-success transition-transform duration-200 group-hover:scale-110"
                            />
                          ) : (
                            <Circle
                              size={20}
                              weight="regular"
                              className="text-muted-foreground transition-all duration-200 group-hover:text-primary group-hover:scale-110"
                            />
                          )}
                        </button>
                        <div className="flex-1 min-w-0">
                          <p
                            className={cn(
                              "text-sm leading-relaxed transition-all duration-200",
                              finding.completed
                                ? "line-through text-muted-foreground"
                                : "text-foreground",
                            )}
                          >
                            {finding.text}
                          </p>
                        </div>
                        <div className="shrink-0 flex items-center gap-2">
                          {severityIcon(finding.severity)}
                          <Badge
                            variant={severityVariant(finding.severity)}
                            className="text-[0.65rem] capitalize"
                          >
                            {finding.severity}
                          </Badge>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteFinding(finding.id);
                            }}
                            className="opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:text-destructive cursor-pointer"
                          >
                            <Trash size={14} />
                          </button>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="p-0">
                <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <CheckCircle
                    size={40}
                    weight="duotone"
                    className="mb-3 opacity-40"
                  />
                  <p className="text-sm font-medium">No findings yet</p>
                  <p className="text-xs mt-1 text-muted-foreground/70">
                    Findings will appear here as the probe analyzes the codebase
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
        </motion.div>
      )}
    </motion.div>
  );
}
