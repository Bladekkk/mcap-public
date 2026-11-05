'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Progress } from '@/components/ui/progress'

const javaVersions = ['Java 8', 'Java 11', 'Java 17', 'Java 21']
const serverTypes = ['Paper', 'Purpur', 'Vanilla', 'Fabric', 'Forge', 'Quilt']
const memoryOptions = [512, 1024, 2048, 4096, 8192]

export default function CreateServerPage() {
  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  const [progress, setProgress] = useState(0)
  
  const [formData, setFormData] = useState({
    name: '',
    jarType: 'Paper',
    version: '',
    javaVersion: 'Java 17',
    xms: 2048,
    xmx: 4096,
    port: 25565,
    eula: false,
  })

  const handleNext = () => {
    setStep((prev) => Math.min(prev + 1, 5))
  }

  const handleBack = () => {
    setStep((prev) => Math.max(prev - 1, 1))
  }

  const handleSubmit = async () => {
    setLoading(true)
    try {
      const res = await fetch('/api/v1/servers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })
      if (res.ok) {
        window.location.href = '/servers'
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container mx-auto p-4 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Create Server</h1>
        <Link href="/servers">
          <Button variant="outline">Cancel</Button>
        </Link>
      </div>

      <div className="mb-8">
        {[1, 2, 3, 4, 5].map((s) => (
          <span
            key={s}
            className={`inline-block h-2 w-2 rounded-full mx-1 ${
              s <= step ? 'bg-primary' : 'bg-muted'
            }`}
          />
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>
            {step === 1 && 'Server Configuration'}
            {step === 2 && 'Java Settings'}
            {step === 3 && 'Resource Allocation'}
            {step === 4 && 'Review'}
            {step === 5 && 'Installing...'}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {step === 1 && (
            <>
              <div className="space-y-2">
                <Label htmlFor="name">Server Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="my-server"
                />
              </div>
              <div className="space-y-2">
                <Label>Server Type</Label>
                <Select
                  value={formData.jarType}
                  onValueChange={(v) => setFormData({ ...formData, jarType: v })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {serverTypes.map((t) => (
                      <SelectItem key={t} value={t}>{t}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Version</Label>
                <Input
                  value={formData.version}
                  onChange={(e) => setFormData({ ...formData, version: e.target.value })}
                  placeholder="1.20.4"
                />
              </div>
            </>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Java Version</Label>
                <Select
                  value={formData.javaVersion}
                  onValueChange={(v) => setFormData({ ...formData, javaVersion: v })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {javaVersions.map((v) => (
                      <SelectItem key={v} value={v}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-6">
              <div className="space-y-2">
                <Label>Memory (Xms/Xmx)</Label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="xms" className="text-xs">Min (MB)</Label>
                    <Input
                      id="xms"
                      type="number"
                      value={formData.xms}
                      onChange={(e) => setFormData({ ...formData, xms: parseInt(e.target.value) })}
                    />
                  </div>
                  <div>
                    <Label htmlFor="xmx" className="text-xs">Max (MB)</Label>
                    <Input
                      id="xmx"
                      type="number"
                      value={formData.xmx}
                      onChange={(e) => setFormData({ ...formData, xmx: parseInt(e.target.value) })}
                    />
                  </div>
                </div>
              </div>
              <div className="space-y-2">
                <Label>Server Port</Label>
                <Input
                  type="number"
                  value={formData.port}
                  onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) })}
                />
              </div>
            </div>
          )}

          {step === 4 && (
            <div className="space-y-4">
              <div className="flex justify-between">
                <span>Name:</span>
                <span className="font-medium">{formData.name}</span>
              </div>
              <div className="flex justify-between">
                <span>Type:</span>
                <span className="font-medium">{formData.jarType}</span>
              </div>
              <div className="flex justify-between">
                <span>Version:</span>
                <span className="font-medium">{formData.version}</span>
              </div>
              <div className="flex justify-between">
                <span>Java:</span>
                <span className="font-medium">{formData.javaVersion}</span>
              </div>
              <div className="flex justify-between">
                <span>Memory:</span>
                <span className="font-medium">{formData.xms} / {formData.xmx} MB</span>
              </div>
              <div className="flex justify-between">
                <span>Port:</span>
                <span className="font-medium">{formData.port}</span>
              </div>
            </div>
          )}

          {step === 5 && (
            <div className="space-y-4">
              <Progress value={progress} className="h-2" />
              <p className="text-sm text-muted-foreground">Installing dependencies...</p>
            </div>
          )}

          <div className="flex justify-between pt-4">
            {step > 1 && step < 5 && (
              <Button variant="outline" onClick={handleBack}>Back</Button>
            )}
            {step < 5 && (
              <Button onClick={handleNext} disabled={!formData.name || !formData.version}>
                Next
              </Button>
            )}
            {step === 5 && (
              <Button disabled={loading}>{loading ? 'Installing...' : 'Install'}</Button>
            )}
            {step === 4 && (
              <Button onClick={handleSubmit} disabled={loading || !formData.eula}>
                {loading ? 'Creating...' : 'Create Server'}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
