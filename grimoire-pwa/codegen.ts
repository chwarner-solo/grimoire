import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: '../grimoire-api/schema/*.graphql',
  documents: 'src/graphql/**/*.graphql',
  generates: {
    'src/graphql/generated/index.ts': {
      plugins: [
        'typescript',
        'typescript-operations',
        'typescript-react-apollo',
      ],
      config: {
        withHooks: true,
        withComponent: false,
        scalars: {
          DateTime: 'string',
          Date: 'string',
        },
      },
    },
  },
}

export default config
