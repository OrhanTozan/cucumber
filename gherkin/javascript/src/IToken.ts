import * as messages from '@cucumber/messages'
import GherkinLine from './GherkinLine'

export default interface IToken<TokenType> {
  location: messages.Location
  line: GherkinLine

  isEof: boolean
  matchedText?: string
  matchedType: TokenType
  matchedItems: GherkinLine[]
  matchedKeyword: string
  matchedIndent: number
  matchedGherkinDialect: string
  getTokenValue(): string
  detach(): void
}
