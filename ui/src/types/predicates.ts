import {Filter, RemoteDataState, TimeRange} from 'src/types'

export interface PredicatesState {
  bucketName: string
  deletionStatus: RemoteDataState
  filters: Filter[]
  isSerious: boolean
  keys: string[]
  timeRange: TimeRange
  values: string[]
}
