"use client"

import * as React from "react"
import { motion } from "framer-motion"
import { Gear, Plus, Trash, FloppyDisk, Key, Globe, Robot } from "@phosphor-icons/react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { getConfig, updateConfig, type Config, type Provider } from "@/lib/api"

const fadeUp = {
  initial: { opacity: 0, y: 12 },
  animate: { opacity: 1, y: 0 },
}

export default function ConfigPage() {
  const [config, setConfig] = React.useState<Config | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [success, setSuccess] = React.useState(false)

  // New provider form
  const [showAdd, setShowAdd] = React.useState(false)
  const [newName, setNewName] = React.useState("")
  const [newBaseURL, setNewBaseURL] = React.useState("")
  const [newAPIKey, setNewAPIKey] = React.useState("")
  const [newModels, setNewModels] = React.useState("")

  React.useEffect(() => {
    loadConfig()
  }, [])

  async function loadConfig() {
    try {
      const data = await getConfig()
      setConfig(data)
    } catch {
      // If API is not available, use empty config
      setConfig({ providers: {}, default: "" })
    } finally {
      setLoading(false)
    }
  }

  async function handleSave() {
    if (!config) return
    setSaving(true)
    setError(null)
    setSuccess(false)
    try {
      await updateConfig(config)
      setSuccess(true)
      setTimeout(() => setSuccess(false), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save config")
    } finally {
      setSaving(false)
    }
  }

  function addProvider() {
    if (!config || !newName.trim()) return

    const models = newModels
      .split(",")
      .map((m) => m.trim())
      .filter(Boolean)

    const provider: Provider = {
      name: newName.trim(),
      base_url: newBaseURL.trim(),
      api_key: newAPIKey.trim(),
      models,
      default_model: models[0] || "",
    }

    const updated = {
      ...config,
      providers: { ...config.providers, [newName.trim()]: provider },
    }

    if (!updated.default) {
      updated.default = newName.trim()
    }

    setConfig(updated)
    setNewName("")
    setNewBaseURL("")
    setNewAPIKey("")
    setNewModels("")
    setShowAdd(false)
  }

  function removeProvider(name: string) {
    if (!config) return
    const providers = { ...config.providers }
    delete providers[name]
    setConfig({
      ...config,
      providers,
      default: config.default === name ? Object.keys(providers)[0] || "" : config.default,
    })
  }

  function setDefault(name: string) {
    if (!config) return
    setConfig({ ...config, default: name })
  }

  function updateProviderField(name: string, field: keyof Provider, value: string) {
    if (!config) return
    const provider = config.providers[name]
    if (!provider) return

    const updated = { ...provider, [field]: value }
    setConfig({
      ...config,
      providers: { ...config.providers, [name]: updated },
    })
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  const providerEntries = config ? Object.entries(config.providers) : []

  return (
    <motion.div
      initial="initial"
      animate="animate"
      transition={{ staggerChildren: 0.06 }}
      className="space-y-8"
    >
      {/* Page header */}
      <motion.div variants={fadeUp} transition={{ duration: 0.3 }} className="space-y-1">
        <div className="flex items-center gap-3">
          <Gear size={24} weight="duotone" className="text-primary" />
          <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
        </div>
        <p className="text-sm text-muted-foreground">
          Manage your AI providers, API keys, and default models.
        </p>
      </motion.div>

      {/* Status messages */}
      {error && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          className="rounded-lg border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
        >
          {error}
        </motion.div>
      )}
      {success && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          className="rounded-lg border border-success/30 bg-success/10 px-4 py-3 text-sm text-success"
        >
          Configuration saved successfully.
        </motion.div>
      )}

      {/* Providers list */}
      <motion.div variants={fadeUp} transition={{ duration: 0.3 }} className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-medium">Providers</h2>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowAdd(!showAdd)}
            className="gap-1.5"
          >
            <Plus size={16} weight="bold" />
            Add Provider
          </Button>
        </div>

        {/* Add provider form */}
        {showAdd && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.2 }}
          >
            <Card className="border-primary/30 border-dashed">
              <CardHeader className="pb-4">
                <CardTitle className="text-base">New Provider</CardTitle>
                <CardDescription>Add a new AI provider configuration.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2">
                    <Label htmlFor="new-name">Name</Label>
                    <Input
                      id="new-name"
                      placeholder="e.g. anthropic"
                      value={newName}
                      onChange={(e) => setNewName(e.target.value)}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="new-url">Base URL</Label>
                    <Input
                      id="new-url"
                      placeholder="https://api.anthropic.com"
                      value={newBaseURL}
                      onChange={(e) => setNewBaseURL(e.target.value)}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="new-key">API Key</Label>
                  <Input
                    id="new-key"
                    type="password"
                    placeholder="sk-..."
                    value={newAPIKey}
                    onChange={(e) => setNewAPIKey(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="new-models">Models (comma-separated)</Label>
                  <Input
                    id="new-models"
                    placeholder="claude-sonnet-4-20250514, claude-3-haiku-20240307"
                    value={newModels}
                    onChange={(e) => setNewModels(e.target.value)}
                  />
                </div>
                <div className="flex gap-2 pt-2">
                  <Button size="sm" onClick={addProvider} disabled={!newName.trim()}>
                    <Plus size={16} weight="bold" />
                    Add
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      setShowAdd(false)
                      setNewName("")
                      setNewBaseURL("")
                      setNewAPIKey("")
                      setNewModels("")
                    }}
                  >
                    Cancel
                  </Button>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}

        {/* Provider cards */}
        {providerEntries.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12 text-center">
              <div className="mb-3 rounded-full bg-muted p-3">
                <Robot size={24} className="text-muted-foreground" />
              </div>
              <p className="text-sm font-medium text-muted-foreground">No providers configured</p>
              <p className="mt-1 text-xs text-muted-foreground">
                Add a provider to get started with code auditing.
              </p>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-3">
            {providerEntries.map(([name, provider], index) => (
              <motion.div
                key={name}
                variants={fadeUp}
                transition={{ duration: 0.3, delay: index * 0.05 }}
              >
                <Card className="transition-shadow hover:shadow-md">
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <CardTitle className="text-base">{name}</CardTitle>
                        {config?.default === name && (
                          <Badge variant="default" className="text-[10px]">
                            Default
                          </Badge>
                        )}
                      </div>
                      <div className="flex items-center gap-1">
                        {config?.default !== name && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setDefault(name)}
                            className="text-xs text-muted-foreground"
                          >
                            Set Default
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => removeProvider(name)}
                          className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        >
                          <Trash size={16} />
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid gap-4 sm:grid-cols-2">
                      <div className="space-y-2">
                        <Label className="flex items-center gap-1.5 text-xs text-muted-foreground">
                          <Globe size={14} />
                          Base URL
                        </Label>
                        <Input
                          value={provider.base_url}
                          onChange={(e) => updateProviderField(name, "base_url", e.target.value)}
                          placeholder="https://api.example.com"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label className="flex items-center gap-1.5 text-xs text-muted-foreground">
                          <Key size={14} />
                          API Key
                        </Label>
                        <Input
                          type="password"
                          value={provider.api_key}
                          onChange={(e) => updateProviderField(name, "api_key", e.target.value)}
                          placeholder="sk-..."
                        />
                      </div>
                    </div>

                    <Separator />

                    <div className="space-y-2">
                      <Label className="flex items-center gap-1.5 text-xs text-muted-foreground">
                        <Robot size={14} />
                        Models
                      </Label>
                      <div className="flex flex-wrap gap-1.5">
                        {provider.models?.map((model) => (
                          <Badge
                            key={model}
                            variant={model === provider.default_model ? "default" : "muted"}
                            className="font-mono text-[11px]"
                          >
                            {model}
                          </Badge>
                        ))}
                        {(!provider.models || provider.models.length === 0) && (
                          <span className="text-xs text-muted-foreground italic">
                            No models configured
                          </span>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        )}
      </motion.div>

      {/* Save button */}
      <motion.div variants={fadeUp} transition={{ duration: 0.3 }}>
        <Separator className="mb-6" />
        <div className="flex items-center justify-end gap-3">
          <Button variant="outline" onClick={loadConfig} disabled={saving}>
            Reset
          </Button>
          <Button onClick={handleSave} disabled={saving} className="gap-1.5">
            <FloppyDisk size={16} weight="bold" />
            {saving ? "Saving..." : "Save Changes"}
          </Button>
        </div>
      </motion.div>
    </motion.div>
  )
}
