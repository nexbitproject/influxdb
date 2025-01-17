/* eslint @typescript-eslint/no-unused-vars: "off" */
import 'jest'

import {
  signin,
  setupUser,
  createDashboard,
  createCell,
  createOrg,
  createSource,
  flush,
  getByTestID,
  getByInputName,
  getByInputValue,
  getByTitle,
  createTask,
  createVariable,
  createMapVariable,
  createAndAddLabel,
  createLabel,
  createBucket,
  createScraper,
  fluxEqual,
  createTelegraf,
  createToken,
  createDashboardTemplate,
  writeData,
  getByTestIDSubStr,
  createEndpoint,
  createDashWithCell,
  createDashWithViewAndVar,
} from './support/commands'

declare global {
  namespace Cypress {
    interface Chainable {
      signin: typeof signin
      setupUser: typeof setupUser
      createSource: typeof createSource
      createTask: typeof createTask
      createVariable: typeof createVariable
      createMapVariable: typeof createMapVariable
      createDashboardTemplate: typeof createDashboardTemplate
      createDashboard: typeof createDashboard
      createCell: typeof createCell
      createDashWithCell: typeof createDashWithCell
      createDashWithViewAndVar: typeof createDashWithViewAndVar
      createOrg: typeof createOrg
      flush: typeof flush
      getByTestID: typeof getByTestID
      getByInputName: typeof getByInputName
      getByInputValue: typeof getByInputValue
      getByTitle: typeof getByTitle
      getByTestIDSubStr: typeof getByTestIDSubStr
      createAndAddLabel: typeof createAndAddLabel
      createLabel: typeof createLabel
      createBucket: typeof createBucket
      createScraper: typeof createScraper
      fluxEqual: typeof fluxEqual
      createTelegraf: typeof createTelegraf
      createToken: typeof createToken
      writeData: typeof writeData
      createEndpoint: typeof createEndpoint
    }
  }
}
