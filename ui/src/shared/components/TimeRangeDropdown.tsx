// Libraries
import React, {PureComponent, createRef} from 'react'
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

// Constants
import {
  SELECTABLE_TIME_RANGES,
  CUSTOM_TIME_RANGE_LABEL,
  TIME_RANGE_FORMAT,
} from 'src/shared/constants/timeRanges'

// Types
import {TimeRange} from 'src/types'

export enum RangeType {
  Absolute = 'absolute',
  Relative = 'relative',
}

interface Props {
  timeRange: TimeRange
  onSetTimeRange: (timeRange: TimeRange) => void
}

interface State {
  isDatePickerOpen: boolean
}

class TimeRangeDropdown extends PureComponent<Props, State> {
  private dropdownRef = createRef<HTMLDivElement>()

  constructor(props: Props) {
    super(props)

    this.state = {isDatePickerOpen: false}
  }

  public render() {
    const timeRange = this.timeRange
    return (
      <>
        <Popover
          type={PopoverType.Outline}
          position={PopoverPosition.ToTheLeft}
          triggerRef={this.dropdownRef}
          visible={this.state.isDatePickerOpen}
          showEvent={PopoverInteraction.None}
          hideEvent={PopoverInteraction.None}
          distanceFromTrigger={8}
          testID="timerange-popover"
          enableDefaultStyles={false}
          contents={() => (
            <DateRangePicker
              timeRange={timeRange}
              onSetTimeRange={this.handleApplyTimeRange}
              onClose={this.handleHideDatePicker}
              position={
                this.state.isDatePickerOpen ? {position: 'relative'} : undefined
              }
            />
          )}
        />
        <div ref={this.dropdownRef}>
          <Dropdown
            style={{width: `${this.dropdownWidth}px`}}
            testID="timerange-dropdown"
            button={(active, onClick) => (
              <Dropdown.Button active={active} onClick={onClick}>
                {timeRange.label}
              </Dropdown.Button>
            )}
            menu={onCollapse => (
              <Dropdown.Menu
                onCollapse={onCollapse}
                style={{width: `${this.dropdownWidth + 50}px`}}
              >
                <Dropdown.Divider
                  key="Time Range"
                  text="Time Range"
                  id="Time Range"
                />
                <Dropdown.Item
                  key={CUSTOM_TIME_RANGE_LABEL}
                  value={CUSTOM_TIME_RANGE_LABEL}
                  id={CUSTOM_TIME_RANGE_LABEL}
                  testID="dropdown-item-custom"
                  selected={this.state.isDatePickerOpen}
                  onClick={this.handleClickCustomTimeRange}
                >
                  {CUSTOM_TIME_RANGE_LABEL}
                </Dropdown.Item>
                {SELECTABLE_TIME_RANGES.map(({label}) => {
                  const testID = label.toLowerCase().replace(/\s/g, '')
                  return (
                    <Dropdown.Item
                      key={label}
                      value={label}
                      id={label}
                      testID={`dropdown-item-${testID}`}
                      selected={label === timeRange.label}
                      onClick={this.handleClickDropdownItem}
                    >
                      {label}
                    </Dropdown.Item>
                  )
                })}
              </Dropdown.Menu>
            )}
          />
        </div>
      </>
    )
  }

  private get dropdownWidth(): number {
    if (this.props.timeRange.type === 'custom') {
      return 250
    }
    return 100
  }

  private get timeRange(): TimeRange {
    const {timeRange} = this.props
    const {isDatePickerOpen} = this.state

    if (isDatePickerOpen && timeRange.type === 'selectable-duration') {
      //convert selected Time Range to custom

      const upper = new Date().toISOString()

      const lower = moment()
        .subtract(timeRange.seconds, 's')
        .toISOString()

      const label = `${moment(lower).format(TIME_RANGE_FORMAT)} - ${moment(
        upper
      ).format(TIME_RANGE_FORMAT)}`

      return {
        label,
        lower,
        upper,
        type: 'custom',
      }
    }

    if (
      timeRange.type === 'custom' ||
      timeRange.type === 'selectable-duration'
    ) {
      return timeRange
    }

    throw new Error(
      `TimeRangeDropdown passed unknown TimeRange with type: ${
        timeRange.type
      }, lower: ${timeRange.lower}, and upper: ${timeRange.upper}`
    )
  }

  private handleApplyTimeRange = (timeRange: TimeRange) => {
    this.props.onSetTimeRange(timeRange)
    this.handleHideDatePicker()
  }

  private handleHideDatePicker = () => {
    this.setState({isDatePickerOpen: false})
  }

  private handleClickCustomTimeRange = (): void => {
    this.setState({isDatePickerOpen: true})
  }

  private handleClickDropdownItem = (label: string): void => {
    const {onSetTimeRange} = this.props
    const timeRange = SELECTABLE_TIME_RANGES.find(t => t.label === label)

    onSetTimeRange(timeRange)
  }
}

export default TimeRangeDropdown
