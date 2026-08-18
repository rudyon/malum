import observatoryImage from './assets/winter-observatory.svg'
import type { ArticleDocument } from './document'

export const sampleArticle: ArticleDocument = {
  title: "the cartographer's winter",
  author: 'Mara Vale',
  source: {
    label: 'fieldnotes.local',
    url: 'https://fieldnotes.local/cartographers-winter',
  },
  lead: [
    {
      type: 'image',
      src: observatoryImage,
      alt: 'A mountain observatory and its amber-lit map room beneath a winter sky.',
      caption:
        'The north station after the first snow. Its map room occupied the long eastern wing, where the morning light arrived before the heat did.',
    },
    {
      type: 'paragraph',
      text: 'The first map I trusted was wrong in three different ways. The river had moved, the old road had collapsed into a ravine, and the village marked in careful black type had been empty for eleven years. I carried it anyway. A bad map at least tells you what somebody once believed, and in unfamiliar country that can be nearly as useful as the truth.',
    },
    {
      type: 'paragraph',
      text: 'By dusk I had corrected the river in blue pencil, crossed out the road, and written a date beside the village. Then I folded the sheet along its old seams and slept with it under my coat.',
    },
    {
      type: 'paragraph',
      text: 'I have always liked the moment when a place stops being scenery and becomes a system. A ridge explains the wind in the valley. A line of alder trees reveals water beneath frozen ground. Chimney smoke, bent almost flat, tells you which pass will be closed by morning. None of these signs mean much alone. Together they make an argument about the land, and for a little while the argument can feel complete.',
    },
    {
      type: 'paragraph',
      text: 'That winter at the north station, completeness was our most dangerous illusion. We had instruments for pressure, snowfall, temperature, and light. We had eighty years of notebooks and a cabinet of maps drawn by people who knew every footpath by name. What we did not have was a way to notice the things nobody had thought to measure.',
    },
  ],
  sections: [
    {
      id: 'the-station-keepers',
      title: 'The Station Keepers',
      blocks: [
        {
          type: 'paragraph',
          text: 'There were five of us at the station that year: an astronomer who distrusted forecasts, two surveyors who disagreed about everything except tea, a mechanic who could repair a clock by listening to it, and me. We shared the rooms above the archive and divided the night watches according to no system anyone would admit to inventing.',
        },
      ],
    },
    {
      id: 'reading-the-snow',
      title: 'Reading the Snow',
      blocks: [
        {
          type: 'paragraph',
          text: 'Fresh snow makes a clean page only from a distance. Up close it is crowded with revisions: windward faces scoured down to ice, hollows collecting soft drifts, the pinprick trail of a bird landing where no branch appears to be. Each mark describes both an event and everything that happened after it.',
        },
      ],
    },
    {
      id: 'the-blank-valley',
      title: 'The Blank Valley',
      blocks: [
        {
          type: 'heading',
          id: 'a-line-left-open',
          level: 3,
          text: 'A line left open',
        },
        {
          type: 'paragraph',
          text: 'On the oldest survey, the eastern valley ended in an unfinished contour. Later cartographers copied the omission so faithfully that it began to look intentional. We decided to walk there before the deepest snow arrived, carrying more confidence than rope.',
        },
      ],
    },
    {
      id: 'instruments-and-omens',
      title: 'Instruments and Omens',
      blocks: [
        {
          type: 'paragraph',
          text: 'An instrument is a disciplined kind of attention. An omen is attention without the discipline. At the station we pretended the distinction was obvious, then changed our minds whenever the barometer behaved strangely or the ravens gathered on the southern roof.',
        },
      ],
    },
    {
      id: 'the-long-night-watch',
      title: 'The Long Night Watch',
      blocks: [
        {
          type: 'paragraph',
          text: 'During the longest nights, the work contracted to small rituals: clear the ice from the gauge, wind the clock, record the temperature, shade the moon on the chart. Repetition made the hours manageable and made every deviation impossible to ignore.',
        },
      ],
    },
    {
      id: 'errors-in-blue-pencil',
      title: 'Errors in Blue Pencil',
      blocks: [
        {
          type: 'paragraph',
          text: 'We used blue for water and for doubt. By January the working map was webbed with both. Some corrections sharpened the country; others merely documented our confusion with greater precision.',
        },
      ],
    },
    {
      id: 'what-the-river-remembered',
      title: 'What the River Remembered',
      blocks: [
        {
          type: 'paragraph',
          text: 'Below the ice, the river kept its own record. Gravel banks appeared where the current had slowed, whole trees lodged at bends, and dark seams in the cut banks marked floods beyond living memory. The landscape had an archive; it simply did not arrange itself for readers.',
        },
      ],
    },
    {
      id: 'a-map-for-leaving',
      title: 'A Map for Leaving',
      blocks: [
        {
          type: 'paragraph',
          text: 'In spring we made a final clean copy. It was accurate, legible, and already becoming obsolete. I left the blue working sheets in the cabinet beside it. Someone arriving later would need to know not only where we had ended, but how often we had been wrong on the way there.',
        },
      ],
    },
    {
      id: 'the-country-beyond-the-page',
      title: 'The Country Beyond the Page',
      blocks: [
        {
          type: 'paragraph',
          text: 'A map is most honest at its edges. The symbols stop, the paper continues for a fraction of an inch, and then the world resumes without explanation. That margin is not a failure of knowledge. It is an invitation to keep looking.',
        },
      ],
    },
  ],
}

