<template>
<section>
  <h1 class="title">Wie wird transkribiert?</h1>
  <p>
    Ziel soll es sein einen Bestand an Trainingsdaten aus heterogenen
    Fraktur-Schriften zu haben, die <em>diplomatisch</em> transkribiert
    wurden. Als <em>diplomatisch</em> wird hier ganz grundlegend folgendes
    verstanden:
  </p>
  <ul>
    <li>Möglichst nah am Druckbild</li>
    <li>Möglichst wenig Interpretation bei der Transkription</li>
    <li>Druckfehler werden nicht korrigiert</li>
  </ul>
  <p>
  Um die Eingabe von auf heutigen Tastaturen nicht mehr vorhandenen Zeichen zu
  vereinfachen steht bei der Transkriptiono eine virtuelle Tastatur zur
  Verfügung die die gebräuchlichsten Zeichen zugänglich macht.
  </p>

  <h2 class="subtitle">Ligaturen: Immer aufspalten (außer <strong>ß</strong>)</h2>
  <p>
  <strong>Ligaturen</strong> wie <FrakturLetter letter="tz" /> (tz) werden <strong>immer aufgespalten</strong>.
  Ausgenommen hiervon ist lediglich <strong><FrakturLetter letter="ß" /> (ß), dieses
  wird immer als solches transkribiert.</strong>
  </p>
  <table class="sample">
    <tr>
      <td><img class="sample" src="../../img/doc/ligature_sample1.png" title="entgegengeſetzt" /></td>
      <td>entgegengeſe<strong>tz</strong>t</td>
    </tr>
    <tr>
      <td><img class="sample" src="../../img/doc/ligature_sample2.png" title="Verhängniß" /></td>
      <td>Verhängni<strong>ß</strong></td>
    </tr>
  </table>

  <h2 class="subtitle"><FrakturLetter letter="j_upper" />: Immer als <strong>J</strong></h2>
  <p>
  In den meisten Frakturschriften gibt es für <strong>I</strong> und <strong>J</strong>
  nur ein einziges Zeichen, <FrakturLetter letter="J_upper" />. Dieses wird
  ausschließlich vorlagengetreu <strong>als J transkribiert</strong>, auch
  in Abkürzungen.
  </p>
  <table class="sample">
    <tr>
      <td><img class="sample" src="../../img/doc/ij_sample1.png" title="Jnneren" /></td>
      <td><strong>J</strong>nneren</td>
    </tr>
    <tr>
      <td><img class="sample" src="../../img/doc/ij_sample2.png" title="Bardenwerper, J. L." /></td>
      <td>Bardenwerper, <strong>J</strong>. L.</td>
    </tr>
  </table>

  <h2 class="subtitle"><FrakturLetter letter="s_long" />: Immer als <strong>ſ</strong></h2>
  <p>
  Sowohl in Fraktur- als auch in Antiquatexten können zwei Formen des
  Kleinbuchstabens <strong>s</strong> auftreten: <FrakturLetter letter="s_long" />
  und <FrakturLetter letter="s_lower" />. Sie werden unterschieden, d.h.
  <FrakturLetter letter="s_long" /> wird als <strong>ſ</strong> transkribiert,
  nicht als s.
  </p>
  <table class="sample">
    <tr>
      <td><img class="sample" src="../../img/doc/s_sample1.png" title="Seeſchiffen" /></td>
      <td>See<strong>ſ</strong>chiffen</td>
    </tr>
    <tr>
      <td><img class="sample" src="../../img/doc/s_sample2.png" title="Buſſeto."/></td>
      <td>Bu<strong>ſſ</strong>eto.</td>
    </tr>
  </table>

  <h2 class="subtitle"><FrakturLetter letter="r_round" />: Immer als <strong>ꝛ</strong></h2>
  <p>
  Gelegentlich wird das "runde r" <FrakturLetter letter="r_round" /> verwendet,
  meist nach einem noch heute gebräuchlichen <em>r</em> wie in <em>Herr</em> oder noch
  häufiger als Substituent für <em>et</em> in <em>et cetera</em> Abkürzungen.
  Es wird als solches transkribiert und nicht mit <em>r</em> substituiert.
  </p>
  <table class="sample">
    <tr>
      <td><img class="sample" src="../../img/doc/r_sample1.png" title="Berge ꝛc." /></td>
      <td>Berge <strong>ꝛ</strong>c.</td>
    </tr>
  </table>

  <h2 class="subtitle">Alte Umlautformen <FrakturLetter letter="umlauts_small" />: Nicht normalisieren</h2>
  <p>
  Inbesondere in Texten der ersten Hälfte des 19. Jahrhunderts finden sich
  noch sehr häufig ältere Formen der heute gebräuchlichen Umlaute <em>äöüÄÖÜ</em>,
  in denen statt den Umlautpunkten ein kleines <em>e</em> verwendet wird:
  <FrakturLetter letter="umlauts_small" />. Diese werden auch so transkribiert
  und nicht auf ihre heute gebräuchliche Form normalisiert.
  </p>
  <div class="notification is-warning">
    Unter Umständen zeigt der Browser diese Zeichen nicht korrekt an.
    Sie sind auf der virtuellen Tastatur in der ersten Zeile zu finden.
  </div>
  <table class="sample">
    <tr>
      <td><img class="sample" src="../../img/doc/umlauts_sample1.png" title="Zubeho&#868;ungen."/></td>
      <td>Zubeh<strong>o&#868;</strong>rungen.</td>
    </tr>
    <tr>
      <td><img class="sample" src="../../img/doc/umlauts_sample2.png" title="troͤ&#868;teſten"/></td>
      <td>tr<strong>o&#868;</strong>ſten</td>
    </tr>
  </table>


  <h2 class="subtitle">Sonstiges</h2>
  <ul>
    <li>Der Doppelbindestrich <strong>⸗</strong> wird als solcher transkribiert</li>
    <li>Doppelte Anführungszeichen wie <strong>„…“</strong>, <strong>“…”</strong>, etc.
      werden als <strong>„…“</strong> normalisiert, d.h. die Ausrichtung der Zeichen
      ist irrelevant und es wird die reguläre deutsche Variante transkribiert.</li>
    <li><strong>»‥«</strong> wird als solches transkribiert und nicht normalisiert</li>
  </ul>

  <hr />

  <form @submit.prevent="finish">
    <div class="field">
      <b-checkbox v-model="hasRead">
        <strong>Ich habe die Richtlinien gelesen und werde versuchen sie einzuhalten</strong>
      </b-checkbox>
    </div>
    <button v-if="hasRead" class="button is-success"><b-icon icon="check" /><span>Los geht's!</span></button>
  </form>
</section>
</template>

<script>
import { mapMutations, mapState } from 'vuex'
import FrakturLetter from './FrakturLetter'

export default {
  name: 'Guidelines',
  components: { FrakturLetter },
  data () {
    return {
      hasRead: false
    }
  },
  computed: mapState(['previousSession']),
  methods: {
    finish () {
      localStorage.setItem('hasReadTheGuidelines', 'true')
      if (this.previousSession) {
        this.changeScreen('restore')
      } else {
        this.changeScreen('config')
      }
    },
    ...mapMutations(['changeScreen'])
  }
}
</script>

<style>
img.sample {
  height: 2em;
}

h2.subtitle {
  margin-top: 2em;
}

.fraktur-letter {
  margin: unset;
}

.fraktur-letter > img {
  height: 1em;
  filter: none;
}

.sample tr td:nth-child(1) {
  text-align: right;
}
.sample tr td:nth-child(2) {
  text-align: left;
  font-family: monospace;
}
</style>
