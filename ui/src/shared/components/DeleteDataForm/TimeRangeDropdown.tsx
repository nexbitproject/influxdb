// Libraries
import React, {useRef, useState, FC} from 'react'
import moment from 'moment'

// Components
import {
  Dropdown,
  Popover,
  PopoverPosition,
  PopoverInteraction,
  PopoverType,
} from '@influxdata/clockface'
import DateRangePicker from 'src/shared/components/dateRangePicker/DateRangePicker'

// Types
import {TimeRange} from 'src/types'

interface Props {
  timeRange: TimeRange
  onSetTimeRange: (timeRange: TimeRange) => void
}

const TimeRangeDropdown: FC<Props> = ({timeRange, onSetTimeRange}) => {
  const [pickerActive, setPickerActive] = useState(false)
  const buttonRef = useRef<HTMLDivElement>(null)

  const readableLower = moment(timeRange.lower).format('YYYY-MM-DD HH:mm:ss')
  const readableUpper = moment(timeRange.upper).format('YYYY-MM-DD HH:mm:ss')

  const handleApplyTimeRange = (timeRange: TimeRange) => {
    onSetTimeRange(timeRange)
    setPickerActive(false)
  }

  return (
    <div ref={buttonRef}>
      <Dropdown.Button onClick={() => setPickerActive(!pickerActive)}>
        {readableLower} - {readableUpper}
      </Dropdown.Button>
      <Popover
        type={PopoverType.Outline}
        position={PopoverPosition.Below}
        triggerRef={buttonRef}
        visible={pickerActive}
        showEvent={PopoverInteraction.None}
        hideEvent={PopoverInteraction.None}
        distanceFromTrigger={8}
        testID="timerange-popover"
        enableDefaultStyles={false}
        contents={() => (
          <DateRangePicker
            timeRange={timeRange}
            onSetTimeRange={handleApplyTimeRange}
            onClose={() => setPickerActive(false)}
            position={{position: 'relative'}}
          />
        )}
      />
    </div>
  )
}

export default TimeRangeDropdown
