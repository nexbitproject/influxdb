import {Filter, RemoteDataState, CustomTimeRange} from 'src/types'

export interface PredicatesState {
  bucketName: string
  deletionStatus: RemoteDataState
  filters: Filter[]
  isSerious: boolean
  keys: string[]
  timeRange: CustomTimeRange
  values: string[]
}
