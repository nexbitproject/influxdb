// Libraries
import React, {FC} from 'react'
import {connect} from 'react-redux'
import {Form, Grid, Columns, Panel} from '@influxdata/clockface'

// Components
import BucketsDropdown from 'src/shared/components/DeleteDataForm/BucketsDropdown'
import TimeRangeDropdown from 'src/shared/components/DeleteDataForm/TimeRangeDropdown'
import Checkbox from 'src/shared/components/Checkbox'
import DeleteButton from 'src/shared/components/DeleteDataForm/DeleteButton'
import FilterEditor from 'src/shared/components/DeleteDataForm/FilterEditor'

// Types
import {Filter, RemoteDataState, CustomTimeRange, AppState} from 'src/types'

// Selectors
import {setCanDelete} from 'src/shared/selectors/canDelete'

// Actions
import {
  deleteFilter,
  deleteWithPredicate,
  resetFilters,
  setFilter,
  setIsSerious,
  setBucketAndKeys,
  setTimeRange,
} from 'src/shared/actions/predicates'

interface OwnProps {
  handleDismiss: () => void
}

interface StateProps {
  bucketName: string
  canDelete: boolean
  deletionStatus: RemoteDataState
  filters: Filter[]
  isSerious: boolean
  keys: string[]
  timeRange: CustomTimeRange
  values: (string | number)[]
}

interface DispatchProps {
  deleteFilter: typeof deleteFilter
  deleteWithPredicate: typeof deleteWithPredicate
  resetFilters: typeof resetFilters
  setFilter: typeof setFilter
  setIsSerious: typeof setIsSerious
  setBucketAndKeys: typeof setBucketAndKeys
  setTimeRange: typeof setTimeRange
}

export type Props = StateProps & DispatchProps & OwnProps

const DeleteDataForm: FC<Props> = ({
  canDelete,
  deleteFilter,
  deletionStatus,
  deleteWithPredicate,
  filters,
  handleDismiss,
  isSerious,
  keys,
  resetFilters,
  setFilter,
  setIsSerious,
  setBucketAndKeys,
  setTimeRange,
  timeRange,
  values,
}) => {
  const handleDelete = () => {
    deleteWithPredicate()
    handleDismiss()
  }

  const handleBucketClick = (selectedBucketName: string) => {
    setBucketAndKeys(selectedBucketName)
    resetFilters()
  }

  return (
    <Form className="delete-data-form">
      <Grid>
        <Grid.Row>
          <Grid.Column widthXS={Columns.Four}>
            <Form.Element label="Target Bucket">
              <BucketsDropdown
                bucketName={name}
                onSetBucketName={handleBucketClick}
              />
            </Form.Element>
          </Grid.Column>
          <Grid.Column widthXS={Columns.Eight}>
            <Form.Element label="Time Range">
              <TimeRangeDropdown
                timeRange={timeRange}
                onSetTimeRange={setTimeRange}
              />
            </Form.Element>
          </Grid.Column>
        </Grid.Row>
        <Grid.Row>
          <Grid.Column widthXS={Columns.Twelve}>
            <FilterEditor
              bucket={name}
              filters={filters}
              keys={keys}
              onDeleteFilter={deleteFilter}
              onSetFilter={setFilter}
              shouldValidate={isSerious}
              values={values}
            />
          </Grid.Column>
        </Grid.Row>
        <Grid.Row>
          <Grid.Column widthXS={Columns.Twelve}>
            <Panel className="delete-data-form--danger-zone">
              <Panel.Header>
                <Panel.Title>Danger Zone!</Panel.Title>
              </Panel.Header>
              <Panel.Body className="delete-data-form--confirm">
                <Checkbox
                  testID="delete-checkbox"
                  label="I understand that this cannot be undone."
                  checked={isSerious}
                  onSetChecked={isSerious => setIsSerious(isSerious)}
                />
                <DeleteButton
                  status={deletionStatus}
                  valid={canDelete}
                  onClick={handleDelete}
                />
              </Panel.Body>
            </Panel>
          </Grid.Column>
        </Grid.Row>
      </Grid>
    </Form>
  )
}

const mstp = ({predicates}: AppState): StateProps => {
  const {
    bucketName,
    deletionStatus,
    filters,
    isSerious,
    keys,
    timeRange,
    values,
  } = predicates
  return {
    bucketName,
    canDelete: setCanDelete(predicates),
    deletionStatus,
    filters,
    isSerious,
    keys,
    timeRange,
    values,
  }
}

const mdtp: DispatchProps = {
  deleteFilter,
  deleteWithPredicate,
  resetFilters,
  setFilter,
  setIsSerious,
  setBucketAndKeys,
  setTimeRange,
}

export default connect<StateProps, DispatchProps>(
  mstp,
  mdtp
)(DeleteDataForm)
