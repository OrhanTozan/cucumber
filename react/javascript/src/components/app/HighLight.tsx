import React from 'react'
import SearchQueryContext from '../../SearchQueryContext'
import elasticlunr from 'elasticlunr'
import highlightWords from 'highlight-words'
import ReactMarkdown from 'react-markdown'
import rehypeRaw from 'rehype-raw'
import sanitizerGithubSchema from 'hast-util-sanitize/lib/github.json'
import rehypeSanitize from 'rehype-sanitize'

interface IProps {
  text: string
  markdown?: boolean
  className?: string
}

const allQueryWords = (queryWords: string[]): string[] => {
  return queryWords.reduce((allWords, word) => {
    const stem = elasticlunr.stemmer(word)
    allWords.push(word)

    if (stem !== word) {
      allWords.push(stem)
    }
    return allWords
  }, [] as string[])
}

const HighLight: React.FunctionComponent<IProps> = ({ text, markdown = false, className = '' }) => {
  const searchQueryContext = React.useContext(SearchQueryContext)
  const query = allQueryWords(
    searchQueryContext.query ? searchQueryContext.query.split(' ') : []
  ).join(' ')
  const appliedClassName = className ? `highlight ${className}` : 'highlight'

  const chunks = highlightWords({ text, query })

  if (markdown) {
    const highlightedText = chunks
      .map(({ text, match }) => (match ? `<mark>${text}</mark>` : text))
      .join('')

    const sanitizerSchema = sanitizerGithubSchema

    sanitizerSchema['tagNames'].push('section')
    sanitizerSchema['attributes']['*'].push('className')

    return (
      <div className={appliedClassName}>
        <ReactMarkdown rehypePlugins={[rehypeRaw, [rehypeSanitize, sanitizerSchema]]}>
          {highlightedText}
        </ReactMarkdown>
      </div>
    )
  }

  return (
    <span className={appliedClassName}>
      {chunks.map(({ text, match, key }) => (match ? <mark key={key}>{text}</mark> : text))}
    </span>
  )
}

export default HighLight
