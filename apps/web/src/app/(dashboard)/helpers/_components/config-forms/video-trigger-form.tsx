'use client'

// Schema: see schemas.ts > videoTriggerSchema

import { FormTextField, FormRadioGroup, FormSelect, FormSwitch, InfoBanner } from './form-fields'
import type { ConfigFormProps } from './types'

const videoSourceOptions = [
  { value: 'wistia', label: 'Wistia' },
  { value: 'youtube', label: 'YouTube' },
  { value: 'vimeo', label: 'Vimeo' },
]

const youtubeSizeOptions = [
  { value: '560x315', label: '560 x 315' },
  { value: '640x360', label: '640 x 360' },
  { value: '853x480', label: '853 x 480' },
  { value: '1280x720', label: '1280 x 720' },
]

export function VideoTriggerForm({ config, onChange, disabled }: ConfigFormProps) {
  const videoSource = (config.videoSource as string) || 'youtube'
  const videoId = (config.videoId as string) || ''
  const embedType = (config.embedType as string) || 'inline'
  const youtubeSize = (config.youtubeSize as string) || '640x360'
  const autoplay = (config.autoplay as boolean) ?? false
  const showControls = (config.showControls as boolean) ?? true

  const updateConfig = (updates: Record<string, unknown>) => {
    onChange({ ...config, ...updates })
  }

  return (
    <div className="space-y-4">
      <FormRadioGroup
        label="Video Source"
        value={videoSource}
        onValueChange={(value) => updateConfig({ videoSource: value })}
        options={videoSourceOptions}
        disabled={disabled}
      />

      <FormTextField
        label={videoSource === 'wistia' ? 'Wistia Media ID' : videoSource === 'youtube' ? 'YouTube Video ID' : 'Vimeo Video ID'}
        placeholder={`Enter ${videoSource} video ID`}
        value={videoId}
        onChange={(e) => updateConfig({ videoId: e.target.value })}
        disabled={disabled}
      />

      {videoSource === 'wistia' && (
        <FormRadioGroup
          label="Embed Type"
          value={embedType}
          onValueChange={(value) => updateConfig({ embedType: value })}
          options={[
            { value: 'inline', label: 'Inline' },
            { value: 'popover', label: 'Popover' },
          ]}
          disabled={disabled}
        />
      )}

      {videoSource === 'youtube' && (
        <>
          <FormSelect
            label="Video Size"
            value={youtubeSize}
            onValueChange={(value) => updateConfig({ youtubeSize: value })}
            options={youtubeSizeOptions}
            disabled={disabled}
          />
          <FormSwitch
            label="Show player controls"
            checked={showControls}
            onCheckedChange={(checked) => updateConfig({ showControls: checked })}
            disabled={disabled}
          />
          <FormSwitch
            label="Autoplay video"
            checked={autoplay}
            onCheckedChange={(checked) => updateConfig({ autoplay: checked })}
            disabled={disabled}
          />
        </>
      )}

      <InfoBanner>
        Set up watch-time triggers using the scoring rules builder to apply tags or trigger goals when viewers reach specific timestamps.
      </InfoBanner>
    </div>
  )
}
