// Libraries
import React, {PureComponent} from 'react'

import {SelectDropdown} from '@influxdata/clockface'
import {BuilderAggregateFunctionType} from 'src/client'

interface Props {
  title: string
  testID: string
  onSelect?: (option: BuilderAggregateFunctionType) => void
  onDelete?: () => void
  onDragStart?: () => void
}

const emptyFunction = () => {}

export default class BuilderCardHeader extends PureComponent<Props> {
  public static defaultProps = {
    testID: 'builder-card--header',
  }

  public render() {
    const {children, onSelect, testID, title} = this.props

    return (
      <div
        className="builder-card--header"
        data-testid={testID}
      >
      <SelectDropdown
        options={['filter', 'group']}
        selectedOption={title}
        testID="select-option-dropdown"
        onSelect={onSelect ? onSelect : emptyFunction}
      />

        {children}
        {this.deleteButton}
      </div>
    )
  }

  private get title(): JSX.Element {
    const {onDragStart, title} = this.props

    if (onDragStart) {
      return (
        <div className="builder-card--draggable" onDragStart={onDragStart}>
          <div className="builder-card--hamburger" />
          <h2 className="builder-card--title">{title}</h2>
        </div>
      )
    }

    return <h2 className="builder-card--title">{title}</h2>
  }

  private get deleteButton(): JSX.Element | undefined {
    const {onDelete} = this.props

    if (onDelete) {
      return <div className="builder-card--delete" onClick={onDelete} />
    }
  }
}
